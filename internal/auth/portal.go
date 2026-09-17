package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Probe checks whether we have WAN access or are trapped in a captive portal.
// Returns isCaptive = true and challengeToken if intercepted.
// Returns isCaptive = false and token = "" if already online (204).
func Probe(ctx context.Context, client *http.Client) (isCaptive bool, challengeToken string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://connectivitycheck.gstatic.com/generate_204", nil)
	if err != nil {
		return false, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 204 {
		return false, "", nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}
	re := regexp.MustCompile(`fgtauth\?([a-f0-9]+)`)

	matches := re.FindStringSubmatch(string(body))

	if len(matches) < 2 {
		return false, "", fmt.Errorf("failed to extract magic token from portal HTML")
	}

	magicToken := matches[1]
	return true, magicToken, nil
}

// Prime registers the challenge session on FortiOS.
func Prime(ctx context.Context, client *http.Client, challengeToken string) error {
	primeUrl := fmt.Sprintf("https://fw.bits-pilani.ac.in:8090/fgtauth?%s", challengeToken)

	req, err := http.NewRequestWithContext(ctx, "GET", primeUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Login posts the credentials and returns the authenticated session token.
func Login(ctx context.Context, client *http.Client, username, password, challengeToken string) (sessionToken string, err error) {
	formData := url.Values{}
	formData.Set("4Tredir", "http://connectivitycheck.gstatic.com/generate_204")
	formData.Set("magic", challengeToken)
	formData.Set("username", username)
	formData.Set("password", password)

	loginUrl := "https://fw.bits-pilani.ac.in:8090/"
	loginReq, err := http.NewRequestWithContext(ctx, "POST", loginUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", err
	}
	primeUrl := fmt.Sprintf("https://fw.bits-pilani.ac.in:8090/fgtauth?%s", challengeToken)

	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("User-Agent", "Mozilla/5.0")
	loginReq.Header.Set("Referer", primeUrl)

	loginResp, err := client.Do(loginReq)
	if err != nil {
		return "", err
	}
	defer loginResp.Body.Close()

	loginBody, err := io.ReadAll(loginResp.Body)
	if err != nil {
		return "", err
	}
	// 1. Extract session token from the keepalive redirect string
	reSession := regexp.MustCompile(`keepalive\?([a-f0-9]+)`)
	sessionMatches := reSession.FindStringSubmatch(string(loginBody))
	if len(sessionMatches) < 2 {
		return "", fmt.Errorf("[ERROR] authentication failed: invalid credentials or session rejected (no keepalive token in response)")
	}
	sessionToken = sessionMatches[1]

	return sessionToken, nil
}

// Logout terminates the session on the gateway.
func Logout(ctx context.Context, client *http.Client, sessionToken string) error {
	logoutURL := fmt.Sprintf("https://fw.bits-pilani.ac.in:8090/logout?%s", sessionToken)
	logoutReq, err := http.NewRequestWithContext(ctx, "GET", logoutURL, nil)
	if err != nil {
		return err
	}
	logoutReq.Header.Set("User-Agent", "Mozilla/5.0")

	logoutResp, err := client.Do(logoutReq)
	if err != nil {
		return err
	}
	defer logoutResp.Body.Close()

	if logoutResp.StatusCode != 200 {
		return fmt.Errorf("logout failed with status code: %d", logoutResp.StatusCode)
	}

	return nil
}

// Keepalive pings the gateway to maintain an active lease.
func Keepalive(ctx context.Context, client *http.Client, sessionToken string) error {
	keepaliveURL := fmt.Sprintf("https://fw.bits-pilani.ac.in:8090/keepalive?%s", sessionToken)
	req, err := http.NewRequestWithContext(ctx, "GET", keepaliveURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("keepalive failed with status code: %d", resp.StatusCode)
	}

	return nil
}
