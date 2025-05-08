package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
)

func Start(version string) error {
	var asJson bool
	var asYaml bool

	rootCmd := &cobra.Command{
		Use:          "runapp",
		SilenceUsage: true,
		Short:        "Run and manage background processes (apps)",
		RunE: func(_ *cobra.Command, _ []string) error {
			list, err := apps.List()
			if err != nil {
				return err
			}

			if asJson {
				b, err := json.Marshal(list)
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}

			if asYaml {
				b, err := yaml.Marshal(list)
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}

			if len(list) == 0 {
				fmt.Println(common.NoAppsMessage)
				return nil
			}

			t := table.New(os.Stdout)

			t.SetHeaders("Name", "Status", "Mode", "PID")
			t.SetHeaderStyle(table.StyleBold)
			t.SetLineStyle(table.StyleBlue)
			t.SetDividers(table.UnicodeRoundedDividers)

			for _, app := range list {
				t.AddRow(app.Name, formatStatus(app.Status, app.ExitCode), common.PrettyRunMode[app.Mode], strconv.Itoa(app.PID))
			}

			t.Render()
			return nil
		},
	}
	rootCmd.Flags().BoolVar(&asJson, "json", false, "output as JSON")
	rootCmd.Flags().BoolVar(&asYaml, "yaml", false, "output as YAML")
	rootCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	rootCmd.AddCommand(buildVersionCmd(version))

	rootCmd.AddCommand(buildRunCmd())
	rootCmd.AddCommand(buildRestartCmd())
	rootCmd.AddCommand(buildLogsCmd())
	rootCmd.AddCommand(buildStatusCmd())

	rootCmd.AddCommand(buildKillCmd())

	rootCmd.AddCommand(buildRemoveCmd())
	rootCmd.AddCommand(buildRemoveManyCmd())

	rootCmd.AddCommand(buildBackgroundCmd())

	rootCmd.AddCommand(buildOnBootCmd())
	rootCmd.AddCommand(buildInstallOnBootCmd())

	return rootCmd.Execute()
}
