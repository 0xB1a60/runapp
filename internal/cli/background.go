package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/gotidy/ptr"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/util"
)

// Go does not natively support fork so we let's get creative
func buildBackgroundCmd() *cobra.Command {
	var appName string

	cmd := &cobra.Command{
		Use:                "background",
		DisableAutoGenTag:  true,
		Hidden:             true,
		DisableSuggestions: true,
		SilenceUsage:       true,
		RunE: func(cobra *cobra.Command, _ []string) error {
			app, err := apps.Get(appName)
			if err != nil {
				return err
			}

			app.Status = common.AppStatusRunning
			app.PID = os.Getpid()
			app.ExitCode = nil
			app.StartedAt = ptr.Of(time.Now())

			if err := app.SaveToFile(); err != nil {
				fmt.Println("error saving cfg", err)
				writeStdErr(app.StderrPath, err)
				return err
			}

			stdoutFile, err := os.Create(app.StdoutPath)
			if err != nil {
				writeStdErr(app.StderrPath, err)
				return err
			}
			defer func(stdoutFile *os.File) {
				if err := stdoutFile.Close(); err != nil {
					fmt.Println("failed to close stderr", err)
				}
			}(stdoutFile)

			stderrFile, err := os.Create(app.StderrPath)
			if err != nil {
				writeStdErr(app.StderrPath, err)
				return err
			}
			defer func(stderrFile *os.File) {
				if err := stderrFile.Close(); err != nil {
					fmt.Println("failed to close stderr", err)
				}
			}(stderrFile)

			args := append(util.GetShellArgs(), app.Command)

			cmd := exec.Command(args[0], args[1:]...)
			cmd.Env = app.Env
			cmd.Dir = app.CWD
			cmd.Stdout = stdoutFile
			cmd.Stderr = stderrFile

			if err := cmd.Start(); err != nil {
				app.ExitCode = ptr.Of(255)
				app.Status = common.AppStatusFailed
				app.FinishedAt = ptr.Of(time.Now())

				if err := app.SaveToFile(); err != nil {
					writeStdErr(app.StderrPath, err)
					return err
				}

				writeStdErr(app.StderrPath, err)
				return err
			}

			done := make(chan error, 1)

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

			killed := false
			go func() {
				if err := cmd.Wait(); err != nil {
					var exitError *exec.ExitError
					if errors.As(err, &exitError) {
						if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
							if !killed {
								app.ExitCode = ptr.Of(status.ExitStatus())
								app.Status = common.AppStatusFailed
								app.FinishedAt = ptr.Of(time.Now())

								if err := app.SaveToFile(); err != nil {
									writeStdErr(app.StderrPath, err)
									done <- err
								}
							}
						}
					}

					writeStdErr(app.StderrPath, err)
					if killed {
						setKilledStatus(app)
					}
					done <- err
					return
				}

				app.ExitCode = ptr.Of(0)
				app.Status = common.AppStatusSuccess
				app.FinishedAt = ptr.Of(time.Now())

				if err := app.SaveToFile(); err != nil {
					writeStdErr(app.StderrPath, err)
				}
				done <- nil
			}()

			ctx := cobra.Context()

			for {
				select {
				case s := <-sig:
					util.DebugLog("signal received: %s", s.String())
					killed = true
					if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
						fmt.Println("failed to send SIGTERM", err)
					}
				case <-done:
					return nil
				case <-ctx.Done():
					killed = true
					if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
						fmt.Println("failed to send SIGTERM", err)
					}
				}
			}

		},
	}
	cmd.PersistentFlags().StringVar(&appName, "name", "", "")
	_ /*ignored as it's not reachable*/ = cmd.MarkPersistentFlagRequired("name")
	return cmd
}

func writeStdErr(path string, err error) {
	stderrFile, stdErr := os.Create(path)
	if stdErr != nil {
		fmt.Println("failed to open stderr", stdErr)
		return
	}
	defer func(stderrFile *os.File) {
		if err := stderrFile.Close(); err != nil {
			fmt.Println("failed to close stderr", err)
		}
	}(stderrFile)

	if _, err := stderrFile.Write([]byte(err.Error())); err != nil {
		fmt.Println("failed to write stderr", err)
	}
}

func setKilledStatus(app *apps.App) {
	app.ExitCode = ptr.Of(137)
	app.Status = common.AppStatusFailed
	app.FinishedAt = ptr.Of(time.Now())

	if err := app.SaveToFile(); err != nil {
		writeStdErr(app.StderrPath, err)
	}
}
