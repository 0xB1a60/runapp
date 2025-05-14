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
	var skipLogs bool

	cmd := &cobra.Command{
		Use:          "restart",
		SilenceUsage: true,
		Short:        "Restart an app",
		Args:         cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var appName string
			if len(args) != 0 {
				appName = args[0]
			}

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
				return errors.New(tml.Sprintf("app is running and cannot be restarted. Use <magenta>runapp kill %s</magenta> to stop it", app.Name))
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
			return viewLogs(cmd.Context(), *app)
		},
	}
	cmd.Flags().BoolVar(&skipLogs, "skip-logs", false, "skip logs streaming after restart")
	return cmd
}

func runApp(app apps.App) error {
	cmd := exec.Command(os.Args[0], "background", app.Name)
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
