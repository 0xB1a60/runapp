package cli

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/require"

	"github.com/0xB1a60/runapp/internal/common"
)

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		name     string
		val      common.AppStatus
		exitCode *int
		expected string
	}{
		{
			name:     "AppStatusFailed without exitCode",
			val:      common.AppStatusFailed,
			exitCode: nil,
			expected: "\x1b[0m\x1b[31mFailed\x1b[39m\x1b[0m",
		},
		{
			name:     "AppStatusFailed with exitCode",
			val:      common.AppStatusFailed,
			exitCode: ptr.Of(1),
			expected: "\x1b[0m\x1b[31mFailed\x1b[39m\x1b[0m (1)",
		},
		{
			name:     "AppStatusSuccess without exitCode",
			val:      common.AppStatusSuccess,
			exitCode: nil,
			expected: "\x1b[0m\x1b[32mSuccess\x1b[39m\x1b[0m",
		},
		{
			name:     "AppStatusSuccess with exitCode",
			val:      common.AppStatusSuccess,
			exitCode: ptr.Of(0),
			expected: "\x1b[0m\x1b[32mSuccess\x1b[39m\x1b[0m (0)",
		},
		{
			name:     "AppStatusRunning without exitCode",
			val:      common.AppStatusRunning,
			exitCode: nil,
			expected: "\x1b[0m\x1b[33mRunning\x1b[39m\x1b[0m",
		},
		{
			name:     "AppStatusStarting without exitCode",
			val:      common.AppStatusStarting,
			exitCode: nil,
			expected: "\x1b[0m\x1b[33mStarting\x1b[39m\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatus(tt.val, tt.exitCode)
			require.Equal(t, tt.expected, result)
		})
	}
}
