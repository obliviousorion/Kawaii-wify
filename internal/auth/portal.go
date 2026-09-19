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

// Pre-compiled regex singletons (allocated once at package init)
var (
	reFgtAuth   = regexp.MustCompile(`fgtauth\?([a-f0-9]+)`)
	reKeepalive = regexp.MustCompile(`keepalive\?([a-f0-9]+)`)
)

// Gateway encapsulates FortiOS captive portal communication, endpoint addresses, and HTTP transport.
type Gateway struct {
	endpoint string       // e.g. "fw.bits-pilani.ac.in:8090"
	host     string       // e.g. "fw.bits-pilani.ac.in"
	client   *http.Client
}

// NewGateway creates a Gateway instance configured for a specific endpoint and TLS host.
func NewGateway(endpoint, host string) *Gateway {
	return &Gateway{
		endpoint: endpoint,
		host:     host,
		client:   NewClient(host),
	}
}

// Endpoint returns the host:port address of the FortiOS captive portal.
func (g *Gateway) Endpoint() string {
	return g.endpoint
}

// Host returns the bare hostname or IP of the gateway.
func (g *Gateway) Host() string {
	return g.host
}

// PrimeURL generates the challenge priming URL for a given challenge token.
func (g *Gateway) PrimeURL(challengeToken string) string {
	return fmt.Sprintf("https://%s/fgtauth?%s", g.endpoint, challengeToken)
}

// LoginURL generates the POST target URL for submitting credentials.
func (g *Gateway) LoginURL() string {
	return fmt.Sprintf("https://%s/", g.endpoint)
}

// KeepaliveURL generates the session keepalive ping URL.
func (g *Gateway) KeepaliveURL(sessionToken string) string {
	return fmt.Sprintf("https://%s/keepalive?%s", g.endpoint, sessionToken)
}

// LogoutURL generates the session termination URL.
func (g *Gateway) LogoutURL(sessionToken string) string {
	return fmt.Sprintf("https://%s/logout?%s", g.endpoint, sessionToken)
}

// Probe checks whether we have WAN access or are trapped in a captive portal.
// Returns isCaptive = true and challengeToken if intercepted.
// Returns isCaptive = false and challengeToken = "" if already online (204).
func (g *Gateway) Probe(ctx context.Context) (isCaptive bool, challengeToken string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://connectivitycheck.gstatic.com/generate_204", nil)
	if err != nil {
		return false, "", err
	}

	resp, err := g.client.Do(req)
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

	matches := reFgtAuth.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return false, "", fmt.Errorf("failed to extract magic token from portal HTML")
	}

	return true, matches[1], nil
}

// Prime registers the challenge session on FortiOS.
func (g *Gateway) Prime(ctx context.Context, challengeToken string) error {
	primeURL := g.PrimeURL(challengeToken)

	req, err := http.NewRequestWithContext(ctx, "GET", primeURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Login posts the credentials and returns the authenticated session token.
func (g *Gateway) Login(ctx context.Context, username, password, challengeToken string) (sessionToken string, err error) {
	formData := url.Values{}
	formData.Set("4Tredir", "http://connectivitycheck.gstatic.com/generate_204")
	formData.Set("magic", challengeToken)
	formData.Set("username", username)
	formData.Set("password", password)

	loginReq, err := http.NewRequestWithContext(ctx, "POST", g.LoginURL(), strings.NewReader(formData.Encode()))
	if err != nil {
		return "", err
	}

	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("User-Agent", "Mozilla/5.0")
	loginReq.Header.Set("Referer", g.PrimeURL(challengeToken))

	loginResp, err := g.client.Do(loginReq)
	if err != nil {
		return "", err
	}
	defer loginResp.Body.Close()

	loginBody, err := io.ReadAll(loginResp.Body)
	if err != nil {
		return "", err
	}

	sessionMatches := reKeepalive.FindStringSubmatch(string(loginBody))
	if len(sessionMatches) < 2 {
		return "", fmt.Errorf("authentication failed: invalid credentials or session rejected (no keepalive token in response)")
	}

	return sessionMatches[1], nil
}

// Logout terminates the session on the gateway.
func (g *Gateway) Logout(ctx context.Context, sessionToken string) error {
	logoutReq, err := http.NewRequestWithContext(ctx, "GET", g.LogoutURL(sessionToken), nil)
	if err != nil {
		return err
	}
	logoutReq.Header.Set("User-Agent", "Mozilla/5.0")

	logoutResp, err := g.client.Do(logoutReq)
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
func (g *Gateway) Keepalive(ctx context.Context, sessionToken string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", g.KeepaliveURL(sessionToken), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("keepalive failed with status code: %d", resp.StatusCode)
	}

	return nil
}
