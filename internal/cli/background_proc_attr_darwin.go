//go:build darwin

package cli

import "syscall"

func backgroundSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
