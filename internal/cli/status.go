package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aquasecurity/table"
	"github.com/liamg/tml"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/tui"
)

func buildStatusCmd() *cobra.Command {
	var asJson bool
	var asYaml bool

	cmd := &cobra.Command{
		Use:          "status",
		SilenceUsage: true,
		Short:        "Read the status of an app",
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
				if asJson || asYaml {
					res := map[string]string{
						"error": "no_apps",
					}

					if asJson {
						b, err := json.Marshal(res)
						if err != nil {
							return err
						}
						fmt.Println(string(b))
						os.Exit(1)
					}

					b, err := yaml.Marshal(res)
					if err != nil {
						return err
					}
					fmt.Println(string(b))
					os.Exit(1)
				}
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
				if asJson || asYaml {
					res := map[string]string{
						"error": "not_found",
					}

					if asJson {
						b, err := json.Marshal(res)
						if err != nil {
							return err
						}
						fmt.Println(string(b))
						os.Exit(1)
					}

					b, err := yaml.Marshal(res)
					if err != nil {
						return err
					}
					fmt.Println(string(b))
					os.Exit(1)
				}

				if errors.Is(err, apps.ErrNotFound) {
					return fmt.Errorf("app: %s does not exist", appName)
				}
				return err
			}

			if asJson {
				b, err := json.Marshal(app)
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}

			if asYaml {
				b, err := yaml.Marshal(app)
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}

			t := table.New(os.Stdout)
			t.SetHeaderStyle(table.StyleBold)
			t.SetLineStyle(table.StyleBlue)
			t.SetDividers(table.UnicodeRoundedDividers)

			t.AddRow("Name", app.Name)
			t.AddRow("Status", formatStatus(app.Status, app.ExitCode))
			t.AddRow("Mode", common.PrettyRunMode[app.Mode])
			t.AddRow("PID", strconv.Itoa(app.PID))
			if app.StartedAt != nil {
				t.AddRow("Started at", app.StartedAt.Format(time.RFC1123))
			}
			if app.FinishedAt != nil {
				t.AddRow("Finished at", app.FinishedAt.Format(time.RFC1123))
			}
			t.AddRow("Command", app.Command)
			t.AddRow("CWD", app.CWD)
			t.AddRow("Stdout", app.StdoutPath)
			t.AddRow("Stderr", app.StderrPath)
			t.AddRow("Env", formatEnv(app.Env))

			t.Render()

			return nil
		},
	}

	cmd.Flags().BoolVar(&asJson, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&asYaml, "yaml", false, "output as YAML")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	return cmd
}

func formatEnv(values []string) string {
	res := ""
	for _, value := range values {
		parts := strings.Split(value, "=")
		res += tml.Sprintf("<cyan>%s</cyan> = %s\n", parts[0], strings.Join(parts[1:], "="))
	}
	return res
}
