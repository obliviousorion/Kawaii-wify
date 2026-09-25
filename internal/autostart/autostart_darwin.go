package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type darwinManager struct{}

func newPlatformManager() (Manager, error) {
	return &darwinManager{}, nil
}

func (m *darwinManager) getPlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "LaunchAgents")
	return filepath.Join(dir, "com.obliviousorion.kawaii-wify.plist"), nil
}

func (m *darwinManager) getExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func (m *darwinManager) Enable() error {
	plistPath, err := m.getPlistPath()
	if err != nil {
		return fmt.Errorf("failed to resolve LaunchAgent path: %w", err)
	}

	exePath, err := m.getExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.obliviousorion.kawaii-wify</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>ProcessType</key>
    <string>Background</string>
</dict>
</plist>
`, exePath)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write LaunchAgent plist: %w", err)
	}

	_ = exec.Command("launchctl", "unload", plistPath).Run()
	if err := exec.Command("launchctl", "load", "-w", plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load LaunchAgent: %w", err)
	}

	return nil
}

func (m *darwinManager) Disable() error {
	plistPath, err := m.getPlistPath()
	if err != nil {
		return fmt.Errorf("failed to resolve LaunchAgent path: %w", err)
	}

	_ = exec.Command("launchctl", "unload", "-w", plistPath).Run()

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove LaunchAgent plist: %w", err)
	}

	return nil
}

func (m *darwinManager) IsEnabled() (bool, error) {
	plistPath, err := m.getPlistPath()
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(plistPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
