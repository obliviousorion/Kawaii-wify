package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestClient_CertificatePinning(t *testing.T) {
	// Start TLS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server url: %v", err)
	}

	serverEndpoint := u.Host
	serverCert := server.TLS.Certificates[0]
	hash := sha256.Sum256(serverCert.Certificate[0])
	validPin := "SHA256:" + strings.ToUpper(hex.EncodeToString(hash[:]))
	invalidPin := "SHA256:00112233445566778899AABBCCDDEEFF00112233445566778899AABBCCDDEEFF"

	t.Run("Valid Pin Matches", func(t *testing.T) {
		client := NewClientWithOptions(ClientOptions{
			TargetEndpoint: serverEndpoint,
			TargetHost:     u.Hostname(),
			GetPins: func(endpoint string) []string {
				return []string{validPin}
			},
			VerifyTLS: true,
		})

		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("expected request to succeed with valid pin, got error: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Mismatched Pin Rejected", func(t *testing.T) {
		client := NewClientWithOptions(ClientOptions{
			TargetEndpoint: serverEndpoint,
			TargetHost:     u.Hostname(),
			GetPins: func(endpoint string) []string {
				return []string{invalidPin}
			},
			VerifyTLS: true,
		})

		_, err := client.Get(server.URL)
		if err == nil {
			t.Fatalf("expected request to fail due to pin mismatch, but it succeeded")
		}
		if !errors.Is(err, ErrFingerprintMismatch) {
			t.Fatalf("expected ErrFingerprintMismatch, got: %v", err)
		}
	})

	t.Run("TOFU Captures Pin", func(t *testing.T) {
		var recordedPin string
		client := NewClientWithOptions(ClientOptions{
			TargetEndpoint: serverEndpoint,
			TargetHost:     u.Hostname(),
			GetPins: func(endpoint string) []string {
				return nil // No existing pins
			},
			OnRecordPin: func(fp string) {
				recordedPin = fp
			},
			VerifyTLS: true,
		})

		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("TOFU request failed: %v", err)
		}
		defer resp.Body.Close()

		if !strings.EqualFold(recordedPin, validPin) {
			t.Fatalf("expected recorded pin %s, got %s", validPin, recordedPin)
		}
	})
}

func TestClient_HostRestrictedRedirect(t *testing.T) {
	// Foreign untrusted server
	evilServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("evil"))
	}))
	defer evilServer.Close()

	// Legitimate gateway server
	var gwServer *httptest.Server
	gwServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/safe-redirect":
			http.Redirect(w, r, "/target", http.StatusFound)
		case "/evil-redirect":
			http.Redirect(w, r, evilServer.URL+"/steal", http.StatusFound)
		case "/target":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("safe target"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer gwServer.Close()

	gwURL, _ := url.Parse(gwServer.URL)

	client := NewClientWithOptions(ClientOptions{
		TargetEndpoint: gwURL.Host,
		TargetHost:     gwURL.Hostname(),
		VerifyTLS:      false,
	})

	t.Run("Same-host redirect succeeds", func(t *testing.T) {
		resp, err := client.Get(gwServer.URL + "/safe-redirect")
		if err != nil {
			t.Fatalf("expected same-host redirect to succeed, got: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got: %d", resp.StatusCode)
		}
	})

	t.Run("Foreign redirect rejected", func(t *testing.T) {
		_, err := client.Get(gwServer.URL + "/evil-redirect")
		if err == nil {
			t.Fatalf("expected foreign redirect to be rejected, but it succeeded")
		}
		if !errors.Is(err, ErrRedirectToForeignHost) {
			t.Fatalf("expected ErrRedirectToForeignHost, got: %v", err)
		}
	})
}

func TestGateway_ForeignPortalDetection(t *testing.T) {
	// Simulate hotel/airport captive portal without FortiGate tokens
	hotelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Welcome to Airport Wi-Fi. Click here to login.</body></html>"))
	}))
	defer hotelServer.Close()

	gw := NewGatewayWithOptions(GatewayOptions{
		Endpoint:  "fw.bits-pilani.ac.in:8090",
		Host:      "fw.bits-pilani.ac.in",
		VerifyTLS: false,
	})

	// Override client with one that directs probe requests to hotelServer
	hotelURL, _ := url.Parse(hotelServer.URL)
	gw.client = &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return nil, nil
			},
		},
	}
	// Create request pointing to hotelServer
	req, _ := http.NewRequestWithContext(t.Context(), "GET", hotelServer.URL, nil)
	resp, err := gw.client.Do(req)
	if err != nil {
		t.Fatalf("failed to query hotel server: %v", err)
	}
	defer resp.Body.Close()

	body := "<html><body>Welcome to Airport Wi-Fi</body></html>"
	matches := reFgtAuth.FindStringSubmatch(body)
	if len(matches) >= 2 {
		t.Fatalf("did not expect fgt token in hotel portal")
	}

	errExpected := fmt.Errorf("%w: response does not contain FortiGate challenge token", ErrForeignPortalDetected)
	if !errors.Is(errExpected, ErrForeignPortalDetected) {
		t.Fatalf("expected ErrForeignPortalDetected")
	}
	_ = hotelURL
}
