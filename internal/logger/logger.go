package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/obliviousorion/kawaii-wify/internal/config"
)

type syncWriter struct {
	file *os.File
	out  io.Writer
}

func (w *syncWriter) Write(p []byte) (n int, err error) {
	n, err = w.out.Write(p)
	if w.file != nil {
		_ = w.file.Sync()
	}
	return n, err
}

// Setup initializes dual-write logging to stdout and daemon.log.
// It rotates the previous session's log to daemon.prev.log.
func Setup() (func(), error) {
	logPath, err := config.LogFilePath()
	if err != nil {
		return func() {}, fmt.Errorf("failed to resolve log path: %w", err)
	}

	prevPath := filepath.Join(filepath.Dir(logPath), "daemon.prev.log")

	// Rotate: old session becomes daemon.prev.log
	if _, err := os.Stat(logPath); err == nil {
		_ = os.Remove(prevPath)
		_ = os.Rename(logPath, prevPath)
	}

	// Open clean log file for this session
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return func() {}, fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}

	writer := &syncWriter{
		file: file,
		out:  io.MultiWriter(os.Stdout, file),
	}

	log.SetOutput(writer)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)

	cleanup := func() {
		_ = file.Sync()
		_ = file.Close()
	}

	return cleanup, nil
}
