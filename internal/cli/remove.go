package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/0xB1a60/runapp/internal/util"
	"github.com/charmbracelet/huh"
	"github.com/liamg/tml"
	"github.com/spf13/cobra"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/tui"
)

func buildRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "remove",
		SilenceUsage: true,
		Short:        "Remove an app",
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

			if app.IsRunning() {
				return errors.New(tml.Sprintf("app is running and cannot be removed. Use <magenta>runapp kill %s</magenta> to stop it", app.Name))
			}
			return os.RemoveAll(app.ConfigPath)
		},
	}
	return cmd
}

func buildRemoveManyCmd() *cobra.Command {
	var removeAllFailed bool
	var removeAllSuccess bool

	cmd := &cobra.Command{
		Use:          "removemany",
		SilenceUsage: true,
		Short:        "Remove apps",
		RunE: func(cmd *cobra.Command, _ []string) error {
			has, err := apps.HasAny()
			if err != nil {
				return err
			}

			if !has {
				fmt.Println(common.NoAppsMessage)
				return nil
			}

			list, err := apps.List()
			if err != nil {
				return err
			}

			hasFailed := false
			hasSuccess := false
			for _, app := range list {
				if app.Status == common.AppStatusFailed {
					hasFailed = true
					continue
				}
				if app.Status == common.AppStatusSuccess {
					hasSuccess = true
					continue
				}
			}

			if !hasSuccess && !hasFailed {
				fmt.Println("🤖 No apps that can be removed")
				return nil
			}

			if removeAllFailed || removeAllSuccess {
				for _, app := range list {
					isFailed := removeAllFailed && app.Status == common.AppStatusFailed
					isSuccess := removeAllSuccess && app.Status == common.AppStatusSuccess
					if isFailed || isSuccess {
						if err := os.RemoveAll(app.ConfigPath); err != nil {
							return err
						}
						fmt.Println(tml.Sprintf("<green>app: %s removed</green>", app.Name))
					}
				}
				return nil
			}

			value, err := appStatusCategorySelect(hasSuccess, hasFailed)
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) || strings.Contains(err.Error(), "could not open a new TTY") {
					util.DebugLog("error while selecting app status category: %s", err)
					return nil
				}
				return err
			}

			for _, app := range list {
				if app.Status == *value {
					if err := os.RemoveAll(app.ConfigPath); err != nil {
						return err
					}
					fmt.Println(tml.Sprintf("<green>app: %s removed</green>", app.Name))
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&removeAllFailed, "failed", false, "all failed apps")
	cmd.Flags().BoolVar(&removeAllSuccess, "success", false, "all successful apps")
	return cmd
}

func appStatusCategorySelect(hasSuccess bool, hasFailed bool) (*common.AppStatus, error) {
	options := make([]huh.Option[common.AppStatus], 0, 2)
	if hasFailed {
		options = append(options, huh.NewOption(formatStatus(common.AppStatusFailed, nil), common.AppStatusFailed))
	}
	if hasSuccess {
		options = append(options, huh.NewOption(formatStatus(common.AppStatusSuccess, nil), common.AppStatusSuccess))
	}

	var value common.AppStatus
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[common.AppStatus]().
			Title("Please select app status?").
			Options(options...).
			Value(&value),
	)).WithTheme(huh.ThemeBase())
	if err := form.Run(); err != nil {
		return nil, err
	}
	return &value, nil
}
