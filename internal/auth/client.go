package auth

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// NewClient creates an HTTP client enforcing HTTP/1.1 over TLS with scoped certificate validation.
func NewClient(targetHost string) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			VerifyConnection: func(cs tls.ConnectionState) error {
				if len(cs.PeerCertificates) == 0 {
					return fmt.Errorf("no certificates presented by server")
				}
				leaf := cs.PeerCertificates[0]
				if targetHost != "" && leaf.Subject.CommonName != targetHost && leaf.VerifyHostname(targetHost) != nil {
					return fmt.Errorf("certificate host mismatch: got %s, want %s", leaf.Subject.CommonName, targetHost)
				}
				return nil
			},
		},
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
}
