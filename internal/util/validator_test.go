package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockEmptyString(t *testing.T) {
	validator := BlockEmptyString("username")

	t.Run("should return error when string is empty", func(t *testing.T) {
		err := validator("")
		require.Error(t, err)
		require.EqualError(t, err, "username is required")
	})

	t.Run("should return nil when string is not empty", func(t *testing.T) {
		err := validator("alice")
		require.NoError(t, err)
	})
}
