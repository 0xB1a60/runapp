package util

import (
	"syscall"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/shirou/gopsutil/v4/process"
)

// collectChildren recursively collects all child processes of the given process
func collectChildren(proc *process.Process) ([]*process.Process, error) {
	var allChildren []*process.Process

	children, err := proc.Children()
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		allChildren = append(allChildren, child)

		// Recursively get child's descendants
		descendants, err := collectChildren(child)
		if err != nil {
			return nil, err
		}

		allChildren = append(allChildren, descendants...)
	}

	return allChildren, nil
}

func killProcessAndChildren(pid int, signal syscall.Signal) error {
	root, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}

	children, err := collectChildren(root)
	if err != nil {
		return err
	}

	allProcesses := append(children, root)

	var multiErr error
	for _, proc := range allProcesses {
		if err := proc.SendSignal(signal); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	if multiErr != nil {
		return multiErr
	}

	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			DebugLog("killProcessAndChildren: waiting for pid and children death timeout exceeded")
			return nil
		case <-ticker.C:
			var exist = false
			for _, proc := range allProcesses {
				if PidExists(int(proc.Pid)) {
					exist = true
					break
				}
			}
			if exist {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			return nil
		}
	}
}

// SoftKill sends a SIGTERM signal to the process and its children
func SoftKill(pid int) error {
	return killProcessAndChildren(pid, syscall.SIGTERM)
}

// ForceKill sends a SIGKILL signal to the process and its children
func ForceKill(pid int) error {
	return killProcessAndChildren(pid, syscall.SIGKILL)
}
