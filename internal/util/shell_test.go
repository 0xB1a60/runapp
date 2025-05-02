package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetShellArgs(t *testing.T) {
	tests := []struct {
		name       string
		shellEnv   string
		wantArgs   []string
		unsetShell bool
	}{
		{
			name:     "Known shell - bash",
			shellEnv: "/bin/bash",
			wantArgs: []string{"/bin/bash", "-c"},
		},
		{
			name:     "Known shell - zsh",
			shellEnv: "/usr/bin/zsh",
			wantArgs: []string{"/usr/bin/zsh", "-c"},
		},
		{
			name:     "Unknown shell",
			shellEnv: "/unknown/shell",
			wantArgs: nil,
		},
		{
			name:       "No SHELL env",
			unsetShell: true,
			wantArgs:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Backup original SHELL env
			originalShell := os.Getenv("SHELL")
			defer func(key, value string) {
				require.NoError(t, os.Setenv(key, value))
			}("SHELL", originalShell)

			if tt.unsetShell {
				require.NoError(t, os.Unsetenv("SHELL"))
			} else {
				t.Setenv("SHELL", tt.shellEnv)
			}

			got := GetShellArgs()
			require.Equal(t, tt.wantArgs, got)
			for i := range got {
				if got[i] != tt.wantArgs[i] {
					require.Equal(t, tt.wantArgs, got)
				}
			}
		})
	}
}
