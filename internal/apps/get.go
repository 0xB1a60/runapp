package apps

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/util"
)

var (
	ErrNotFound = errors.New("app does not exist")
)

func Get(name string) (*App, error) {
	homeDir, err := util.HomeDirPath()
	if err != nil {
		return nil, err
	}

	configContent, err := os.ReadFile(path.Join(homeDir, name, common.FileConfig))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var app App
	if err := json.Unmarshal(configContent, &app); err != nil {
		return nil, err
	}
	app.checkAndCorrectStatus()
	return &app, nil
}
