package auth

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ClientOptions configures the hardened TLS HTTP client.
type ClientOptions struct {
	TargetEndpoint string
	TargetHost     string
	GetPins        func(endpoint string) []string
	OnRecordPin    func(fingerprint string)
	VerifyTLS      bool
}

// NewClient creates an HTTP client enforcing HTTP/1.1 over TLS with scoped certificate validation.
// Preserved for backwards compatibility with single host callers.
func NewClient(targetHost string) *http.Client {
	return NewClientWithOptions(ClientOptions{
		TargetEndpoint: targetHost,
		TargetHost:     targetHost,
		VerifyTLS:      true,
	})
}

// NewClientWithOptions creates an HTTP client with SHA-256 pinning and strict host-scoped redirects.
func NewClientWithOptions(opts ClientOptions) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			VerifyConnection: func(cs tls.ConnectionState) error {
				if !opts.VerifyTLS {
					return nil
				}

				if len(cs.PeerCertificates) == 0 {
					return fmt.Errorf("no certificates presented by server")
				}

				leaf := cs.PeerCertificates[0]
				hash := sha256.Sum256(leaf.Raw)
				fp := "SHA256:" + strings.ToUpper(hex.EncodeToString(hash[:]))

				// 1. Check pinned fingerprints if available
				var pins []string
				if opts.GetPins != nil && opts.TargetEndpoint != "" {
					pins = opts.GetPins(opts.TargetEndpoint)
				}

				if len(pins) > 0 {
					for _, trusted := range pins {
						if strings.EqualFold(trusted, fp) {
							return nil
						}
					}
					return fmt.Errorf("%w: presented %s is not in trusted pins for %s", ErrFingerprintMismatch, fp, opts.TargetEndpoint)
				}

				// 2. TOFU (Trust On First Use) mode
				if opts.TargetHost != "" {
					matchesCN := strings.EqualFold(leaf.Subject.CommonName, opts.TargetHost)
					matchesSAN := leaf.VerifyHostname(opts.TargetHost) == nil
					if !matchesCN && !matchesSAN {
						return fmt.Errorf("certificate host mismatch: got CN=%s, want %s", leaf.Subject.CommonName, opts.TargetHost)
					}
				}

				// Record pending fingerprint to be committed once gateway responds validly
				if opts.OnRecordPin != nil {
					opts.OnRecordPin(fp)
				}

				return nil
			},
		},
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}

			destHost := req.URL.Host
			destHostname := req.URL.Hostname()

			if isAllowedRedirect(destHost, destHostname, opts.TargetEndpoint, opts.TargetHost) {
				return nil
			}

			return fmt.Errorf("%w: destination %s does not match gateway %s", ErrRedirectToForeignHost, destHost, opts.TargetEndpoint)
		},
	}

	return client
}

func isAllowedRedirect(destHost, destHostname, targetEndpoint, targetHost string) bool {
	// Strict match on targetEndpoint (e.g. "fw.bits-pilani.ac.in:8090" or "127.0.0.1:54321")
	if targetEndpoint != "" && strings.EqualFold(destHost, targetEndpoint) {
		return true
	}
	// If targetEndpoint did not specify a port, allow hostname match
	if targetEndpoint != "" && !strings.Contains(targetEndpoint, ":") && strings.EqualFold(destHostname, targetEndpoint) {
		return true
	}
	// If targetHost is explicitly specified without port constraint on targetEndpoint
	if targetHost != "" && !strings.Contains(targetEndpoint, ":") && strings.EqualFold(destHostname, targetHost) {
		return true
	}
	// Allow loopback to Google connectivity check URL
	if strings.EqualFold(destHostname, "connectivitycheck.gstatic.com") {
		return true
	}
	return false
}
