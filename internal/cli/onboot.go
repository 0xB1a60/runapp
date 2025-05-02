package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/liamg/tml"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
)

func buildOnBootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "onboot",
		DisableAutoGenTag:  true,
		Hidden:             true,
		DisableSuggestions: true,
		SilenceUsage:       true,
		RunE: func(_ *cobra.Command, _ []string) error {
			list, err := apps.List()
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Println(common.NoAppsMessage)
				return nil
			}

			fmt.Println("Running on-boot")
			for _, app := range list {
				fmt.Println()

				if app.Status == common.AppStatusRunning || app.Status == common.AppStatusStarting {
					fmt.Println("application already running", app.Name)
					continue
				}

				if app.Mode == common.RunModeOnBoot {
					if err := runOnBootApp(app); err != nil {
						fmt.Println(fmt.Sprintf("error while running (%s) on boot:", app.Name), err)
					}
				}
			}
			fmt.Println("Finished on-boot")
			return nil
		},
	}
	return cmd
}

func runOnBootApp(app apps.App) error {
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
