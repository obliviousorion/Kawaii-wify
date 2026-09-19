//go:build windows

package ipc

import (
	"errors"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// windows pipe path syntax
const pipePath = `\\.\pipe\kawaii-wify`

var (
	ErrDaemonAlreadyRunning = errors.New("daemon is already running on this socket")
)

func Listen() (net.Listener, error) {
	// Guard against duplicate daemon instances
	if conn, err := Dial(); err == nil {
		conn.Close()
		return nil, ErrDaemonAlreadyRunning
	}

	return winio.ListenPipe(pipePath, nil)
}

func Dial() (net.Conn, error) {
	timeout := 2 * time.Second
	return winio.DialPipe(pipePath, &timeout)
}



