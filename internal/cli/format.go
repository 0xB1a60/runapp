package cli

import (
	"fmt"

	"github.com/liamg/tml"

	"github.com/0xB1a60/runapp/internal/common"
)

func formatStatus(val common.AppStatus, exitCode *int) string {
	switch val {
	case common.AppStatusFailed:
		status := tml.Sprintf("<red>Failed</red>")
		if exitCode == nil {
			return status
		}
		return fmt.Sprintf("%s (%d)", status, *exitCode)
	case common.AppStatusSuccess:
		status := tml.Sprintf("<green>Success</green>")
		if exitCode == nil {
			return status
		}
		return fmt.Sprintf("%s (%d)", status, *exitCode)
	case common.AppStatusRunning:
		return tml.Sprintf("<yellow>Running</yellow>")
	case common.AppStatusStarting:
		return tml.Sprintf("<yellow>Starting</yellow>")
	}
	panic("unreachable")
}
