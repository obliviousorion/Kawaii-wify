package config

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultGateway = "fw.bits-pilani.ac.in:8090"

type Config struct {
	Username      string              `json:"username"`
	Gateway       string              `json:"gateway,omitempty"`
	CheckInterval string              `json:"check_interval"`
	Keepalive     bool                `json:"keepalive"`
	AutoConnect   bool                `json:"auto_connect"`
	CertPins      map[string][]string `json:"cert_pins,omitempty"` // "endpoint": ["SHA256:..."]
	VerifyTLS     *bool               `json:"verify_tls,omitempty"` // nil or true = verify, false = bypass
	Notifications *bool               `json:"notifications,omitempty"` // nil or true = enabled, false = disabled
}

func Default() *Config {
	verify := true
	notif := true
	return &Config{
		Gateway:       DefaultGateway,
		CheckInterval: "10s",
		Keepalive:     true,
		AutoConnect:   true,
		CertPins:      make(map[string][]string),
		VerifyTLS:     &verify,
		Notifications: &notif,
	}
}

// GatewayEndpoint returns the host:port combination, defaulting port to 8090 if omitted.
func (c *Config) GatewayEndpoint() string {
	gw := strings.TrimSpace(c.Gateway)
	if gw == "" {
		gw = DefaultGateway
	}
	// Strip scheme if present
	gw = strings.TrimPrefix(gw, "https://")
	gw = strings.TrimPrefix(gw, "http://")
	gw = strings.TrimRight(gw, "/")

	if !strings.Contains(gw, ":") {
		return gw + ":8090"
	}
	return gw
}

// GatewayHost returns the pure hostname or IP without port (for TLS certificate verification).
func (c *Config) GatewayHost() string {
	endpoint := c.GatewayEndpoint()
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return endpoint
	}
	return host
}

// GetPins returns the list of trusted certificate SHA-256 fingerprints for an endpoint.
func (c *Config) GetPins(endpoint string) []string {
	if c.CertPins == nil {
		return nil
	}
	return c.CertPins[endpoint]
}

// AddPin registers a trusted SHA-256 fingerprint for a gateway endpoint.
func (c *Config) AddPin(endpoint, fingerprint string) {
	if c.CertPins == nil {
		c.CertPins = make(map[string][]string)
	}
	for _, fp := range c.CertPins[endpoint] {
		if strings.EqualFold(fp, fingerprint) {
			return
		}
	}
	c.CertPins[endpoint] = append(c.CertPins[endpoint], fingerprint)
}

// ClearPins removes all stored certificate pins for a given endpoint.
func (c *Config) ClearPins(endpoint string) {
	if c.CertPins == nil {
		return
	}
	delete(c.CertPins, endpoint)
}

// IsTLSVerificationEnabled returns whether TLS certificate & pin validation is enforced.
func (c *Config) IsTLSVerificationEnabled() bool {
	if c.VerifyTLS == nil {
		return true
	}
	return *c.VerifyTLS
}

// IsNotificationsEnabled returns whether OS desktop toast notifications are enabled.
func (c *Config) IsNotificationsEnabled() bool {
	if c.Notifications == nil {
		return true
	}
	return *c.Notifications
}

// SetNotificationsEnabled sets the notifications preference.
func (c *Config) SetNotificationsEnabled(enabled bool) {
	c.Notifications = &enabled
}

func Load() (*Config, error) {
	config := Default()

	configPath, err := configPath()
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(configBytes, config); err != nil {
		return nil, err
	}

	return config, nil
}

func Save(config *Config) error {
	configPath, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	configJson, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, configJson, 0600); err != nil {
		return err
	}
	return nil
}

func (c *Config) Interval() time.Duration {
	d, err := time.ParseDuration(c.CheckInterval)
	if err != nil || d <= 0 {
		return 10 * time.Second
	}
	return d
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	pathString := filepath.Join(dir, "kawaii-wify", "config.json")
	return pathString, nil
}
