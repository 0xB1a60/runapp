package util

import (
	"errors"
	"syscall"
)

func PidExists(pid int) bool {
	// Signal 0 does not actually send a signal
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true // process exists
	}
	if errors.Is(err, syscall.ESRCH) {
		return false // process does not exist
	}
	return true // some other error (like EPERM), assume it exists
}
