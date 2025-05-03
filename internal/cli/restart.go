package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/liamg/tml"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/tui"
)

func buildRestartCmd() *cobra.Command {
	var appName string
	var skipLogs bool

	cmd := &cobra.Command{
		Use:          "restart",
		SilenceUsage: true,
		Short:        "Restart an app",
		RunE: func(cmd *cobra.Command, _ []string) error {
			has, err := apps.HasAny()
			if err != nil {
				return err
			}

			if !has {
				fmt.Println(common.NoAppsMessage)
				return nil
			}

			entry := tui.FlagOrPromptEntry{
				Value:        appName,
				TUIFunc:      tui.NamePicker,
				ValidateFunc: nameValidateFunc,
				SetFunc: func(value string) {
					appName = value
				},
			}

			if err := tui.ResolveFlagsOrPrompt(entry); err != nil {
				if errors.Is(err, tui.ErrStop) {
					return nil
				}
				return err
			}

			app, err := apps.Get(appName)
			if err != nil {
				if errors.Is(err, apps.ErrNotFound) {
					return fmt.Errorf("app: %s does not exist", appName)
				}
				return err
			}

			if app.IsRunning() {
				return fmt.Errorf("app: %s is already running", appName)
			}

			app.Status = common.AppStatusStarting
			app.ExitCode = nil
			app.PID = -1
			if err := app.SaveToFile(); err != nil {
				return err
			}

			if err := os.Remove(app.StderrPath); err != nil {
				return err
			}

			if err := os.Remove(app.StdoutPath); err != nil {
				return err
			}

			if err := runApp(*app); err != nil {
				return err
			}

			if skipLogs {
				return nil
			}
			return streamLogs(cmd.Context(), *app)
		},
	}
	cmd.PersistentFlags().StringVar(&appName, "name", "", "name of an app")
	cmd.PersistentFlags().BoolVar(&skipLogs, "skip-logs", false, "skip logs streaming after restart")
	return cmd
}

func runApp(app apps.App) error {
	cmd := exec.Command(os.Args[0], "background", "--name", app.Name)
	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // start new session
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	fmt.Println(tml.Sprintf("<italic>%s</italic> started with PID: %d", app.Name, cmd.Process.Pid))
	return nil
}
