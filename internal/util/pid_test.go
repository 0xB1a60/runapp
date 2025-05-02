package util

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPidExists(t *testing.T) {
	t.Run("existing pid should return true", func(t *testing.T) {
		cmd := exec.Command("sleep", "5")
		err := cmd.Start()
		require.NoError(t, err)
		defer func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}()

		pid := cmd.Process.Pid
		exists := PidExists(pid)
		require.True(t, exists, "expected PidExists to return true for running process")
	})

	t.Run("non-existent pid should return false", func(t *testing.T) {
		fakePid := 999999 // unlikely to be real
		require.False(t, PidExists(fakePid), "expected PidExists to return false for nonexistent PID")
	})
}
