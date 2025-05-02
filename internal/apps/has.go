package apps

import (
	"os"

	"github.com/0xB1a60/runapp/internal/util"
)

func HasAny() (bool, error) {
	homeDir, err := util.HomeDirPath()
	if err != nil {
		return false, err
	}

	entries, err := os.ReadDir(homeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return true, nil
		}
	}
	return false, nil
}
