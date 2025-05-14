package util

import (
	"os"
	"path"
)

const (
	runAppDir = "runapp"
)

// GetRunAppDir returns the path to the runapp directory in the user's home directory
func HomeDirPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(homeDir, ".config", runAppDir), nil
}
