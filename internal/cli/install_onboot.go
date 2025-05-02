package cli

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/liamg/tml"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/tui"
	"github.com/0xB1a60/runapp/internal/util"
)

//go:embed onboot/runapp-boot.service.tpl
var systemdSvcTpl string

const (
	mvSvcCmd          = "mv ./runapp-boot.service " + common.SystemdSvcPath
	daemonReloadCmd   = "systemctl --user daemon-reload"
	enableStartSvcCmd = "systemctl --user enable --now runapp-boot.service"
	checkStatusCmd    = "systemctl status --user runapp-boot.service"
)

var commands = []string{
	mvSvcCmd,
	daemonReloadCmd,
	enableStartSvcCmd,
}

const (
	defaultBinaryPath = "/usr/local/bin/runapp"
)

func buildInstallOnBootCmd() *cobra.Command {
	var binaryPath string
	cmd := &cobra.Command{
		Use:          "install-onboot",
		Short:        "Set up a systemd service to automatically start runapp at boot",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			entries := []tui.FlagOrPromptEntry{
				{
					Value:        binaryPath,
					ValidateFunc: util.BlockEmptyString("binary path"),
					TUIFunc:      tui.TextWithPlaceholder("Where is runapp located?", defaultBinaryPath),
					SetFunc: func(value string) {
						binaryPath = value
					},
				},
			}

			if err := tui.ResolveFlagsOrPrompt(entries...); err != nil {
				if errors.Is(err, tui.ErrStop) {
					return nil
				}
				return err
			}

			afterBinaryPath := strings.ReplaceAll(systemdSvcTpl, "$BINARY_PATH", binaryPath)

			systemdPath, err := util.ResolvePath(common.SystemdPath)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(systemdPath, os.ModePerm); err != nil {
				return err
			}

			actionFunc := func(_ context.Context) error {
				systemdSvcPath, err := util.ResolvePath(common.SystemdSvcPath)
				if err != nil {
					return err
				}

				if err := os.WriteFile(systemdSvcPath, []byte(afterBinaryPath), os.ModePerm); err != nil {
					return err
				}

				for _, command := range commands {
					if err := util.ExecuteCommand(command, true); err != nil {
						util.DebugLog("failed to execute on-boot command: %v", err)
					}
				}
				return nil
			}

			err = spinner.New().
				Title("Executing commands...").
				ActionWithErr(actionFunc).
				Run()
			if err != nil {
				util.DebugLog("error in execute spinner: %v", err)

				// fallback to user ran commands
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}

				if err := os.WriteFile(path.Join(cwd, "runapp-boot.service"), []byte(afterBinaryPath), os.ModePerm); err != nil {
					return err
				}

				fmt.Println("Execute the following commands to install the systemd service:")
				fmt.Println(tml.Sprintf("<magenta>%s</magenta>", mvSvcCmd))
				fmt.Println(tml.Sprintf("<magenta>%s</magenta>", daemonReloadCmd))
				fmt.Println(tml.Sprintf("<magenta>%s</magenta>", enableStartSvcCmd))

				fmt.Println("After a few seconds check the status of the systemd service:")
				fmt.Println(tml.Sprintf("<magenta>%s</magenta>", checkStatusCmd))

				if err := util.ExecuteCommand(checkStatusCmd, false); err != nil {
					util.DebugLog("failed to execute on-boot command: %v", err)
				}
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&binaryPath, "binary-path", "", "path to the runapp binary")
	return cmd
}
