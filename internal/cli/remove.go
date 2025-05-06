package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/liamg/tml"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/tui"
)

func buildRemoveCmd() *cobra.Command {
	var appName string
	var removeAllFailed bool
	var removeAllSuccess bool

	cmd := &cobra.Command{
		Use:          "remove",
		SilenceUsage: true,
		Short:        "Remove an app",
		RunE: func(cmd *cobra.Command, _ []string) error {
			has, err := apps.HasAny()
			if err != nil {
				return err
			}

			if !has {
				fmt.Println(common.NoAppsMessage)
				return nil
			}

			if removeAllFailed || removeAllSuccess {
				list, err := apps.List()
				if err != nil {
					return err
				}
				for _, app := range list {
					isFailed := removeAllFailed && app.Status == common.AppStatusFailed
					isSuccess := removeAllSuccess && app.Status == common.AppStatusSuccess
					if isFailed || isSuccess {
						if err := os.RemoveAll(app.ConfigPath); err != nil {
							return err
						}
					}
				}
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
				return errors.New(tml.Sprintf("app is running and cannot be removed. Use <magenta>runapp kill --name %s</magenta> to stop it", app.Name))
			}
			return os.RemoveAll(app.ConfigPath)
		},
	}
	cmd.Flags().StringVar(&appName, "name", "", "name of an app")
	cmd.Flags().BoolVar(&removeAllFailed, "all-failed", false, "all failed apps")
	cmd.Flags().BoolVar(&removeAllSuccess, "all-success", false, "all successful apps")
	cmd.MarkFlagsMutuallyExclusive("all-failed", "name")
	cmd.MarkFlagsMutuallyExclusive("all-success", "name")
	return cmd
}
