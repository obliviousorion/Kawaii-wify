//go:build !windows

package ipc

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

var (
	ErrDaemonAlreadyRunning = errors.New("daemon is already running on this socket")
)

func socketPath() string {
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, "kawaii-wify.sock")
	}

	return filepath.Join(os.TempDir(), "kawaii-wify.sock")
}

func Listen() (net.Listener, error) {
	path := socketPath()

	if _, err := os.Stat(path); err == nil {
		conn, dialErr := net.Dial("unix", path)
		if dialErr == nil {
			conn.Close()
			return nil, ErrDaemonAlreadyRunning
		}

		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("failed to remove stale socket: %w", err)
		}
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("failed to bind unix socket: %w", err)
	}

	if err := os.Chmod(path, 0600); err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("failed to set socket permissions: %w", err)
	}
	return listener, nil
}

func Dial() (net.Conn, error) {
	return net.Dial("unix", socketPath())
}
