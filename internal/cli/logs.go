package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/liamg/tml"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/tui"
	"github.com/0xB1a60/runapp/internal/util"
)

func buildLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "logs",
		SilenceUsage: true,
		Short:        "Stream the logs (stdout,stderr) of an app",
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

			return streamLogs(cmd.Context(), *app)
		},
	}
	return cmd
}

func streamLogs(ctx context.Context, app apps.App) error {
	fmt.Println(tml.Sprintf("<yellow>▶ Streaming logs for app: %s. You can stop the streaming with CTRL+C, the process won't be interrupted</yellow>", app.Name))

	if !app.IsRunning() {
		if err := printLines(os.Stdout, app.StdoutPath, false); err != nil {
			return err
		}
		if err := printLines(os.Stderr, app.StderrPath, true); err != nil {
			return err
		}
		return nil
	}

	logs, err := apps.ReadLogs(ctx, app)
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
				}
			case log := <-logs:
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

func printLines(w io.Writer, filename string, asError bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			util.DebugLog("Failed to close file %s: %s", filename, err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if asError {
			if _, err := fmt.Fprintln(w, tml.Sprintf("<red>%s</red>", scanner.Text())); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintln(w, scanner.Text()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}
	return nil
}
