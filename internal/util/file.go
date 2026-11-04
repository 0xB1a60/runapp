package util

import (
	"os"
	"path/filepath"
	"strings"
)

func FileExists(value string) bool {
	path, err := ResolvePath(value)
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	if err == nil {
		return true // file exists
	}
	if os.IsNotExist(err) {
		return false // file does not exist
	}
	// some other error, maybe permission issue
	return false
}

func WriteFile(name string, data []byte) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	if _, err = f.Write(data); err != nil {
		return err
	}

	if err = f.Sync(); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}
	return nil
}

// ResolvePath resolves a path, expanding the home directory if it starts with "~"
func ResolvePath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return filepath.Clean(path), nil
}
