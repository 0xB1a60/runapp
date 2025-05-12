package apps

import (
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/util"
)

type App struct {
	Name   string           `json:"name" yaml:"name"`
	Mode   common.RunMode   `json:"mode" yaml:"mode"`
	Status common.AppStatus `json:"status" yaml:"status"`

	Command string   `json:"command" yaml:"command"`
	PID     int      `json:"pid" yaml:"pid"`
	CWD     string   `json:"cwd" yaml:"cwd"`
	Env     []string `json:"env" yaml:"env"`

	ConfigPath string `json:"config_path" yaml:"config_path"`
	StdoutPath string `json:"stdout_path" yaml:"stdout_path"`
	StderrPath string `json:"stderr_path" yaml:"stderr_path"`

	StartedAt  *time.Time `json:"started_at" yaml:"started_at"`
	ExitCode   *int       `json:"exit_code" yaml:"exit_code"`
	FinishedAt *time.Time `json:"finished_at" yaml:"finished_at"`
}

func (app *App) SaveToFile() error {
	homeDir, err := util.HomeDirPath()
	if err != nil {
		return err
	}

	b, err := json.Marshal(app)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(homeDir, app.Name, common.FileConfig), b, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// because runapp is daemon-less sometimes processes will be killed from the outside and runapp will show them as running
func (app *App) checkAndCorrectStatus() {
	if app.Status == common.AppStatusRunning && !util.PidExists(app.PID) {
		app.Status = common.AppStatusFailed
		if app.ExitCode != nil && *app.ExitCode == 0 {
			app.Status = common.AppStatusSuccess
		}
		if err := app.SaveToFile(); err != nil {
			util.DebugLog("error saving app to file: %v", err)
		}
	}
}

func (app *App) IsRunning() bool {
	return app.Status == common.AppStatusRunning || app.Status == common.AppStatusStarting
}
