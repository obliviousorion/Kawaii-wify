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
	Username      string `json:"username"`
	Gateway       string `json:"gateway,omitempty"`
	CheckInterval string `json:"check_interval"`
	Keepalive     bool   `json:"keepalive"`
	AutoConnect   bool   `json:"auto_connect"`
}

func Default() *Config {
	return &Config{
		Gateway:       DefaultGateway,
		CheckInterval: "10s",
		Keepalive:     true,
		AutoConnect:   true,
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
