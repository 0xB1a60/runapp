package util

import (
	"os"
	"path"
)

const (
	runAppDir = "runapp"
)

func HomeDirPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(homeDir, ".config", runAppDir), nil
}
