package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func buildVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}
