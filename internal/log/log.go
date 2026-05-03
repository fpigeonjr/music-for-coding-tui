// Package log writes structured debug output to ~/.config/music-for-coding/debug.log.
//
// The log is append-only and capped to the last 500 lines to prevent unbounded
// growth. It is safe for concurrent use.
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxLines = 500

var (
	mu   sync.Mutex
	path string
	initOnce sync.Once
)

func configDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "music-for-coding")
}

func initPath() {
	dir := configDir()
	if dir == "" {
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	path = filepath.Join(dir, "debug.log")
}

func logf(level, format string, v ...interface{}) {
	initOnce.Do(initPath)
	if path == "" {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	msg := fmt.Sprintf(format, v...)
	line := fmt.Sprintf("%s [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), level, msg)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.WriteString(line)
	trimToMax(path)
}

func trimToMax(p string) {
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) <= maxLines+1 {
		return
	}
	// Keep the last maxLines (accounting for trailing empty string after Split)
	start := len(lines) - maxLines - 1
	if start < 0 {
		start = 0
	}
	out := strings.Join(lines[start:], "\n")
	_ = os.WriteFile(p, []byte(out), 0o644)
}

// Debug writes a debug-level line.
func Debug(format string, v ...interface{}) { logf("DEBUG", format, v...) }

// Info writes an info-level line.
func Info(format string, v ...interface{}) { logf("INFO", format, v...) }

// Error writes an error-level line.
func Error(format string, v ...interface{}) { logf("ERROR", format, v...) }

// Path returns the current log file path (empty if unavailable).
func Path() string {
	initOnce.Do(initPath)
	return path
}
