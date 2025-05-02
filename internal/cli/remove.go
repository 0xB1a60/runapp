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
	cmd.PersistentFlags().StringVar(&appName, "name", "", "name of an app")
	return cmd
}
