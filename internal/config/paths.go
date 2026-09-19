package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// LogDir returns the platform standard directory where daemon logs reside.
// Windows: %LOCALAPPDATA%\kawaii-wify\logs
// Linux:   $XDG_STATE_HOME/kawaii-wify/logs or ~/.local/state/kawaii-wify/logs
func LogDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			var err error
			baseDir, err = os.UserConfigDir()
			if err != nil {
				return "", err
			}
		}
	default: // Linux / macOS
		stateHome := os.Getenv("XDG_STATE_HOME")
		if stateHome != "" {
			baseDir = stateHome
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			baseDir = filepath.Join(home, ".local", "state")
		}
	}

	dir := filepath.Join(baseDir, "kawaii-wify", "logs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// LogFilePath returns the absolute path to daemon.log.
func LogFilePath() (string, error) {
	dir, err := LogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "daemon.log"), nil
}
