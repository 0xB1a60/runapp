package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/tui"
	"github.com/0xB1a60/runapp/internal/util"
)

func buildKillCmd() *cobra.Command {
	var appName string

	cmd := &cobra.Command{
		Use:          "kill",
		SilenceUsage: true,
		Short:        "Kill an app",
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
				TUIFunc:      namePicker(),
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

			actionFunc := func() {
				time.Sleep(10 * time.Second)

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
				case <-cmd.Context().Done():
				}
			}

			err = spinner.New().
				Title("Killing app...").
				Action(actionFunc).
				Run()
			if err != nil {
				util.DebugLog("error in kill spinner: %v", err)
				fmt.Println("Killing app....")
				actionFunc()
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&appName, "name", "", "name of app")
	return cmd
}

func namePicker() func() (*string, error) {
	return func() (*string, error) {
		list, err := apps.List()
		if err != nil {
			return nil, err
		}

		options := make([]huh.Option[string], 0, len(list))
		idx := make(map[string]apps.App, len(list))
		for _, app := range list {
			title := app.Name

			if app.Status == common.AppStatusSuccess || app.Status == common.AppStatusFailed {
				title = "💀 " + title
			}
			idx[app.Name] = app
			options = append(options, huh.NewOption(title, app.Name))
		}

		var value string
		form := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("My app is:").
				Options(options...).
				Validate(func(appName string) error {
					if app := idx[appName]; app.Status == common.AppStatusSuccess || app.Status == common.AppStatusFailed {
						return fmt.Errorf("app has already: %s", common.AppStatusPretty[app.Status])
					}
					return nil
				}).
				Value(&value),
		)).WithTheme(huh.ThemeBase())
		if err := form.Run(); err != nil {
			return nil, err
		}
		return &value, nil
	}
}
