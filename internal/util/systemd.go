package util

import (
	"os"
	"strings"
	"sync"
)

var IsSystemd = sync.OnceValue(isSystemdUsed)

func isSystemdUsed() bool {
	// Check PID 1's command name
	data, err := os.ReadFile("/proc/1/comm")
	if err == nil && strings.TrimSpace(string(data)) == "systemd" {
		return true
	}

	// As a fallback, check for systemd private socket
	if _, err := os.Stat("/run/systemd/private"); err == nil {
		return true
	}

	return false
}
