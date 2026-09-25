package autostart

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName    = "KawaiiWify"
)

type windowsManager struct{}

func newPlatformManager() (Manager, error) {
	return &windowsManager{}, nil
}

func (m *windowsManager) Enable() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve executable: %w", err)
	}
	exePath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		exePath = exe
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	cmd := fmt.Sprintf("\"%s\" daemon", exePath)
	if err := key.SetStringValue(appName, cmd); err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	return nil
}

func (m *windowsManager) Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	if err := key.DeleteValue(appName); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	return nil
}

func (m *windowsManager) IsEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appName)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
