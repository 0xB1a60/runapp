package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/liamg/tml"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/tui"
	"github.com/0xB1a60/runapp/internal/util"
)

func buildKillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kill",
		SilenceUsage: true,
		Short:        "Kill an app",
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
				Value: appName,
				TUIFunc: func() (*string, error) {
					return tui.NamePickerWithValidator(namePickerValidator)
				},
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

			if !app.IsRunning() {
				return errors.New("app is not running")
			}

			killApp(cmd.Context(), app)
			fmt.Println(tml.Sprintf("<green>App successfully killed 💀</green>"))
			return nil
		},
	}
	return cmd
}

func killApp(ctx context.Context, app *apps.App) {
	actionFunc := func() {
		done := make(chan error)
		go func() {
			done <- util.SoftKill(app.PID)
		}()

		select {
		case err := <-done:
			if err != nil {
				util.DebugLog("Failed to stop app: %v", err)
			}
			util.DebugLog("app stopped")
		case <-time.After(time.Second * 10):
			util.DebugLog("force killing app")
			if err := util.ForceKill(app.PID); err != nil {
				util.DebugLog("failed to force kill app: %v", err)
			}
			util.DebugLog("app force killed")
		case <-ctx.Done():
		}
	}

	err := spinner.New().
		Title("Killing app...").
		Action(actionFunc).
		Run()
	if err != nil {
		util.DebugLog("error in kill spinner: %v", err)
		fmt.Println("Killing app....")
		actionFunc()
	}
}

func namePickerValidator(list []apps.App, appName string) error {
	var foundApp *apps.App
	for _, app := range list {
		if app.Name == appName {
			foundApp = &app
		}
	}
	if foundApp == nil {
		util.DebugLog("app %s not found", appName)
		return nil
	}

	if foundApp.Status == common.AppStatusSuccess || foundApp.Status == common.AppStatusFailed {
		return errors.New("app is not running")
	}
	return nil
}
