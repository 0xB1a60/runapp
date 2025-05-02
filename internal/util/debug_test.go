package util

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- helper function for resetting the cached debug value ---
func resetCachedDebug() {
	cachedDebug = nil
}

func TestIsDebug(t *testing.T) {
	originalDebug := os.Getenv("DEBUG")
	defer func() { _ = os.Setenv("DEBUG", originalDebug) }()

	t.Run("should return false when DEBUG is not set", func(t *testing.T) {
		resetCachedDebug()
		_ = os.Unsetenv("DEBUG")

		require.False(t, IsDebug(), "expected IsDebug to be false when DEBUG env var is not set")
	})

	t.Run("should return true when DEBUG is 'true'", func(t *testing.T) {
		resetCachedDebug()
		require.NoError(t, os.Setenv("DEBUG", "true"))

		require.True(t, IsDebug(), "expected IsDebug to be true when DEBUG env var is 'true'")
	})

	t.Run("should be case insensitive", func(t *testing.T) {
		resetCachedDebug()
		require.NoError(t, os.Setenv("DEBUG", "TrUe"))

		require.True(t, IsDebug(), "expected IsDebug to be true when DEBUG env var is 'TrUe'")
	})

	t.Run("should return false when DEBUG is 'false'", func(t *testing.T) {
		resetCachedDebug()
		require.NoError(t, os.Setenv("DEBUG", "false"))

		require.False(t, IsDebug(), "expected IsDebug to be false when DEBUG env var is 'false'")
	})
}

func TestDebugLog(t *testing.T) {
	originalDebug := os.Getenv("DEBUG")
	defer func() { _ = os.Setenv("DEBUG", originalDebug) }()

	t.Run("should print when debug is enabled", func(t *testing.T) {
		resetCachedDebug()
		require.NoError(t, os.Setenv("DEBUG", "true"))

		// Capture stdout
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		DebugLog("hello %s", "world")

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		output := buf.String()
		require.Contains(t, output, "hello world", "expected debug log output to contain formatted string")
	})

	t.Run("should not print when debug is disabled", func(t *testing.T) {
		resetCachedDebug()
		require.NoError(t, os.Setenv("DEBUG", "false"))

		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		DebugLog("this should not appear")

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		output := buf.String()
		require.Equal(t, "", strings.TrimSpace(output), "expected no debug log output")
	})
}
