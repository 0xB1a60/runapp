package util

import (
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

// applyToProcessAndChildren applies a function to the given process and all its children
func applyToProcessAndChildren(pid int, applyFunc func(*process.Process) error) error {
	root, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}

	children, err := collectChildren(root)
	if err != nil {
		return err
	}

	var res error
	for _, proc := range append(children, root) {
		if err := applyFunc(proc); err != nil {
			res = multierror.Append(res, err)
		}
	}
	return res
}

// SoftKill sends a SIGTERM signal to the process and its children
func SoftKill(pid int) error {
	return applyToProcessAndChildren(pid, func(proc *process.Process) error {
		return proc.Terminate()
	})
}

// ForceKill sends a SIGKILL signal to the process and its children
func ForceKill(pid int) error {
	return applyToProcessAndChildren(pid, func(proc *process.Process) error {
		return proc.Kill()
	})
}
