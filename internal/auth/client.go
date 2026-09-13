package auth

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

func NewClient() *http.Client {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			VerifyConnection:   verifyConnection,
		},
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
}

func verifyConnection(cs tls.ConnectionState) error {
	if len(cs.PeerCertificates) == 0 {
		return fmt.Errorf("no certificates presented by server")
	}
	leaf := cs.PeerCertificates[0]
	if leaf.Subject.CommonName != "fw.bits-pilani.ac.in" && leaf.VerifyHostname("fw.bits-pilani.ac.in") != nil {
		return fmt.Errorf("certificate host mismatch: %s", leaf.Subject.CommonName)
	}
	return nil
}
