package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type linuxManager struct{}

func newPlatformManager() (Manager, error) {
	return &linuxManager{}, nil
}

func (m *linuxManager) getServicePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "systemd", "user")
	return filepath.Join(dir, "kawaii-wify.service"), nil
}

func (m *linuxManager) getExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func (m *linuxManager) Enable() error {
	servicePath, err := m.getServicePath()
	if err != nil {
		return fmt.Errorf("failed to resolve service path: %w", err)
	}

	exePath, err := m.getExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(servicePath), 0755); err != nil {
		return fmt.Errorf("failed to create systemd user directory: %w", err)
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=Kawaii-Wify FortiGate WLAN Session Manager
After=network.target

[Service]
Type=simple
ExecStart=%s daemon
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`, exePath)

	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service unit file: %w", err)
	}

	if _, err := exec.LookPath("systemctl"); err == nil {
		_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
		if err := exec.Command("systemctl", "--user", "enable", "--now", "kawaii-wify.service").Run(); err != nil {
			return fmt.Errorf("systemctl enable failed: %w", err)
		}
	}

	return nil
}

func (m *linuxManager) Disable() error {
	servicePath, err := m.getServicePath()
	if err != nil {
		return fmt.Errorf("failed to resolve service path: %w", err)
	}

	if _, err := exec.LookPath("systemctl"); err == nil {
		_ = exec.Command("systemctl", "--user", "stop", "kawaii-wify.service").Run()
		_ = exec.Command("systemctl", "--user", "disable", "kawaii-wify.service").Run()
	}

	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service unit file: %w", err)
	}

	if _, err := exec.LookPath("systemctl"); err == nil {
		_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	}

	return nil
}

func (m *linuxManager) IsEnabled() (bool, error) {
	servicePath, err := m.getServicePath()
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(servicePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "--user", "is-enabled", "kawaii-wify.service")
		if err := cmd.Run(); err == nil {
			return true, nil
		}
	}

	return true, nil
}
