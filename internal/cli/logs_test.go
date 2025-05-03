package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		asError  bool
		expected string
	}{
		{
			name:     "normal lines",
			content:  "line1\nline2",
			asError:  false,
			expected: "line1\nline2\n",
		},
		{
			name:     "error lines",
			content:  "err1\nerr2",
			asError:  true,
			expected: "\x1b[0m\x1b[31merr1\x1b[39m\x1b[0m\n\x1b[0m\x1b[31merr2\x1b[39m\x1b[0m",
		},
		{
			name:     "empty file",
			content:  "",
			asError:  false,
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "testfile")
			require.NoError(t, err)
			defer func(name string) {
				require.NoError(t, os.Remove(name))
			}(tmpfile.Name()) // clean up

			_, err = tmpfile.WriteString(tc.content)
			require.NoError(t, err)
			require.NoError(t, tmpfile.Close())

			var buf bytes.Buffer
			require.NoError(t, printLines(&buf, tmpfile.Name(), tc.asError))

			// Strip ANSI color codes to compare raw text if needed
			output := strings.TrimSpace(buf.String())
			expected := strings.TrimSpace(tc.expected)

			require.Equal(t, expected, output)
		})
	}
}

func TestPrintLines_FileNotExist(t *testing.T) {
	var buf bytes.Buffer
	err := printLines(&buf, "nonexistent.txt", false)
	if err == nil || !strings.Contains(err.Error(), "failed to open file") {
		require.NoError(t, err)
	}
}
