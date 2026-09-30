package auth

import "errors"

var (
	// ErrFingerprintMismatch indicates the presented certificate does not match trusted SHA-256 pins.
	ErrFingerprintMismatch = errors.New("TLS certificate fingerprint mismatch: potential MITM interception")

	// ErrForeignPortalDetected indicates an unknown or foreign captive portal intercepted the traffic.
	ErrForeignPortalDetected = errors.New("foreign captive portal detected: redirect target does not match gateway")

	// ErrRedirectToForeignHost indicates an HTTP redirect attempted to steer traffic away from the gateway.
	ErrRedirectToForeignHost = errors.New("refusing redirect to untrusted foreign host")
)
