package ipc

import "time"

// --- Types for Telemetry & Status --- 

// StatusRequest carries optional query parameters for status checks.
type StatusRequest struct{}

// StatusResponse contains the live operational metrics of the daemon.
type StatusResponse struct {
	State        string        `json:"state"`         // "Online", "Captive", "Offline", "Cooldown", "Paused"
	Username     string        `json:"username"`      // Active student ID
	SessionToken string        `json:"session_token"` // FortiGate auth token
	IPAddress    string        `json:"ip_address"`    // Local IP on campus interface
	LastProbe    time.Time     `json:"last_probe"`    // Timestamp of last network check
	Uptime       time.Duration `json:"uptime"`        // Total running duration
	Paused       bool          `json:"paused"`        // True if explicitly paused via disconnect
}

// --- Generic Actions ---

// ActionRequest carries parameters for basic trigger actions.
type ActionRequest struct{}

// ActionResponse reports whether a commanded action succeeded.
type ActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- Credentials & Identity ---

// SetCredentialsRequest sends updated campus credentials to the daemon.
type SetCredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// --- Configuration ---

// ConfigRequest queries the daemon's active operational settings.
type ConfigRequest struct{}

// ConfigResponse returns the active configuration values.
type ConfigResponse struct {
	Username      string `json:"username"`
	CheckInterval string `json:"check_interval"`
	Keepalive     bool   `json:"keepalive"`
}

// UpdateConfigRequest modifies runtime operational parameters.
type UpdateConfigRequest struct {
	CheckInterval string `json:"check_interval"`
	Keepalive     bool   `json:"keepalive"`
}