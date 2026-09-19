//go:build windows

package ipc

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// windows pipe path syntax
const pipePath = `\\.\pipe\kawaii-wify`

func Listen() (net.Listener, error) {
	return winio.ListenPipe(pipePath, nil)
}

func Dial() (net.Conn, error) {
	timeout := 2 * time.Second
	return winio.DialPipe(pipePath, &timeout)
}


