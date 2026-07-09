package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/liamg/tml"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/logs"
	"github.com/0xB1a60/runapp/internal/tui"
	"github.com/0xB1a60/runapp/internal/util"
)

func buildLogsCmd() *cobra.Command {
	var logType logs.LogType

	cmd := &cobra.Command{
		Use:          "logs",
		SilenceUsage: true,
		Short:        "Stream the logs (stdout,stderr) of an app",
		Args:         cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !slices.Contains(logs.ValidTypes, logType) {
				return fmt.Errorf("type must be %s, %s or %s", logs.AllLogs, logs.OutLogs, logs.ErrLogs)
			}

			has, err := apps.HasAny()
			if err != nil {
				return err
			}

			if !has {
				fmt.Println(common.NoAppsMessage)
				return nil
			}

			var appName string
			if len(args) != 0 {
				appName = args[0]
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

			return viewLogs(cmd.Context(), *app, logType)
		},
	}
	cmd.Flags().StringVar(&logType, "type", logs.AllLogs,
		fmt.Sprintf("type of logs to show (one of: %s)", strings.Join(logs.ValidTypes, ", ")))

	return cmd
}

func viewLogs(ctx context.Context, app apps.App, logType logs.LogType) error {
	if !app.IsRunning() {
		return logs.PrintLines(app, logType)
	}

	fmt.Println(tml.Sprintf("<yellow>▶ Streaming logs for app: %s. You can stop the streaming with CTRL+C, the process won't be interrupted</yellow>", app.Name))
	logStream, err := logs.Stream(ctx, app, logType)
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if app, err := apps.Get(app.Name); err == nil {
					if !app.IsRunning() {
						var msg string
						if app.ExitCode == nil {
							if app.Status == common.AppStatusFailed {
								msg = tml.Sprintf("<red>Process completed with status: %s</red>", common.AppStatusPretty[app.Status])
							} else {
								msg = tml.Sprintf("<green>Process completed with status: %s</green>", common.AppStatusPretty[app.Status])
							}
						} else {
							if app.Status == common.AppStatusFailed {
								msg = tml.Sprintf("<red>Process completed with status: %s (%d)</red>", common.AppStatusPretty[app.Status], *app.ExitCode)
							} else {
								msg = tml.Sprintf("<green>Process completed with status: %s (%d)</green>", common.AppStatusPretty[app.Status], *app.ExitCode)
							}
						}
						fmt.Println(msg)
						cancelFunc()
						continue
					}
				} else {
					if errors.Is(err, apps.ErrNotFound) {
						util.DebugLog("app does not exist anymore")
						fmt.Fprintln(os.Stderr, tml.Sprintf("<red>%s</red>", "app was removed"))
						cancelFunc()
						continue
					}
				}
			case log := <-logStream:
				if log.IsErr {
					fmt.Fprintln(os.Stderr, tml.Sprintf("<red>%s</red>", log.Value)) // no lint // handling this error is not needed
					continue
				}
				fmt.Println(log.Value)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	return nil
}
