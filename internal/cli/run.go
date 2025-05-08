package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"syscall"

	"github.com/charmbracelet/huh"
	"github.com/liamg/tml"
	"github.com/moby/moby/pkg/namesgenerator"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/tui"
	"github.com/0xB1a60/runapp/internal/util"
)

func buildRunCmd() *cobra.Command {
	var runOnBoot bool
	var command string
	var skipLogs bool
	var skipSystemdWarning bool

	cmd := &cobra.Command{
		Use:          "run",
		SilenceUsage: true,
		Short:        "Run an app",
		Args:         cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var appName string
			if len(args) != 0 {
				appName = args[0]
			}

			entry := tui.FlagOrPromptEntry{
				Value:        appName,
				TUIFunc:      nameTextInput,
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

			if !runOnBoot {
				if value, err := onBootSelect(); err != nil {
					util.DebugLog("error while building on-boot: %s", err)
				} else {
					runOnBoot = value
				}
			}

			if runOnBoot && !skipSystemdWarning && !util.FileExists(common.SystemdSvcPath) {
				continueWithoutSystemd, err := systemdSvcExistPicker()
				if err != nil {
					util.DebugLog("error while building systemd svc: %s", err)
				} else {
					if !continueWithoutSystemd {
						return nil
					}
				}
			}

			entry = tui.FlagOrPromptEntry{
				Value:        command,
				TUIFunc:      commandText,
				ValidateFunc: commandValidateFunc,
				SetFunc: func(value string) {
					command = value
				},
			}
			if err := tui.ResolveFlagsOrPrompt(entry); err != nil {
				if errors.Is(err, tui.ErrStop) {
					return nil
				}
				return err
			}

			if existingApp, err := apps.Get(appName); err == nil {
				if existingApp.IsRunning() && util.PidExists(existingApp.PID) {
					return errors.New(tml.Sprintf("app is already running. Use <magenta>runapp kill %s</magenta> to stop it", appName))
				}
			}

			runMode := common.RunModeOnce
			if runOnBoot {
				runMode = common.RunModeOnBoot
			}

			return createAndRunApp(cmd.Context(), appName, runMode, command, skipLogs)
		},
	}
	cmd.Flags().BoolVar(&runOnBoot, "start-on-boot", false, "automatically start the app on boot")
	cmd.Flags().BoolVar(&skipLogs, "skip-logs", false, "skip logs streaming after start")
	cmd.Flags().BoolVar(&skipSystemdWarning, "skip-systemd-warning", false, "suppress warning if systemd service is not detected")
	cmd.Flags().StringVar(&command, "command", "", "command that will be executed")
	return cmd
}

const (
	nameMinLength = 2
	nameMaxLength = 100
)

func nameTextInput() (*string, error) {
	placeholder := namesgenerator.GetRandomName(0)

	var value string
	form := huh.NewForm(huh.NewGroup(
		huh.NewText().
			Title("Name:").
			Placeholder(placeholder).
			CharLimit(nameMaxLength).
			Lines(1).
			Validate(nameValidateFunc).
			Value(&value),
	)).WithTheme(huh.ThemeBase())
	if err := form.Run(); err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return &placeholder, nil
	}
	return &value, nil
}

func nameValidateFunc(value string) error {
	if len(value) == 0 {
		return nil
	}

	if len(value) < nameMinLength {
		return fmt.Errorf("name must be at least %d characters long", nameMinLength)
	}

	if len(value) > nameMaxLength {
		return fmt.Errorf("name must be less than %d characters long", nameMaxLength)
	}

	match, _ := regexp.MatchString("^[a-z0-9_-]+$", value)
	if !match {
		return errors.New("name must only contain lowercase, alphanumeric and -_")
	}
	return nil
}

const (
	Yes = "yes"
	No  = "no"
)

func onBootSelect() (bool, error) {
	var value bool
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[bool]().
			Title("Do you want your app to start on boot?").
			Options(
				huh.NewOption(No, false),
				huh.NewOption(Yes, true),
			).
			Value(&value),
	)).WithTheme(huh.ThemeBase())
	if err := form.Run(); err != nil {
		return false, err
	}
	return value, nil
}

func systemdSvcExistPicker() (bool, error) {
	var value bool
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[bool]().
			Title(fmt.Sprintf("Systemd svc: %s does not exist, are you sure want to continue?", common.SystemdSvcPath)).
			Options(
				huh.NewOption(No, false),
				huh.NewOption(Yes, true),
			).
			Value(&value),
	)).WithTheme(huh.ThemeBase())
	if err := form.Run(); err != nil {
		return false, err
	}
	return value, nil
}

func commandText() (*string, error) {
	var value string
	form := huh.NewForm(huh.NewGroup(
		huh.NewText().
			Title("Command:").
			Lines(1).
			Validate(commandValidateFunc).
			Value(&value),
	)).WithTheme(huh.ThemeBase())
	if err := form.Run(); err != nil {
		return nil, err
	}
	return &value, nil
}

func commandValidateFunc(value string) error {
	if len(value) == 0 {
		return errors.New("value must not be empty")
	}
	return nil
}

func createAndRunApp(ctx context.Context, name string, mode common.RunMode, command string, skipLogs bool) error {
	util.DebugLog("Starting: %s with mode: %s and command: %s", name, string(mode), command)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	homeDir, err := util.HomeDirPath()
	if err != nil {
		return err
	}

	runDir := path.Join(homeDir, name)
	if err := os.RemoveAll(runDir); err != nil {
		return err
	}

	if err := os.MkdirAll(runDir, os.ModePerm); err != nil {
		return err
	}

	stdoutPath := path.Join(runDir, common.FileStdOut)
	stdoutFile, err := os.Create(stdoutPath)
	if err != nil {
		return err
	}
	if err := stdoutFile.Close(); err != nil {
		return err
	}

	stdErrPath := path.Join(runDir, common.FileStdErr)
	stderrFile, err := os.Create(stdErrPath)
	if err != nil {
		return err
	}
	if err := stderrFile.Close(); err != nil {
		return err
	}

	app := apps.App{
		Name:       name,
		Mode:       mode,
		Status:     common.AppStatusStarting,
		Command:    command,
		PID:        -1,
		CWD:        cwd,
		Env:        os.Environ(),
		ConfigPath: runDir,
		StderrPath: stdErrPath,
		StdoutPath: stdoutPath,
	}

	if err := app.SaveToFile(); err != nil {
		return err
	}

	cmd := exec.Command(os.Args[0], "background", name)
	cmd.Env = os.Environ()

	if util.IsDebug() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = nil
		cmd.Stderr = nil
	}

	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // start new session
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	fmt.Println(tml.Sprintf("<italic>%s</italic> started with PID: %d", app.Name, cmd.Process.Pid))

	if skipLogs {
		return nil
	}
	return streamLogs(ctx, app)
}
