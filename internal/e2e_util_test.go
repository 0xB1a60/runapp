package internal

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type RunCommandResult struct {
	cmd *exec.Cmd

	stderr []string
	stdout []string

	combined []string
}

func runCommand(command string) (*RunCommandResult, error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Env = os.Environ()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return nil, fmt.Errorf("%d -- %s -- %s -- %s", exitError.ExitCode(), string(exitError.Stderr), stderr.String(), stderr.String())
		}
		return nil, err
	}

	errLines := strings.Split(strings.TrimSpace(stderr.String()), "\n")
	outLines := strings.Split(strings.TrimSpace(stdout.String()), "\n")

	output := make([]string, 0, len(outLines)+len(errLines))
	for _, line := range append(errLines, outLines...) {
		if len(line) != 0 {
			output = append(output, line)
		}
	}

	return &RunCommandResult{
		cmd:      cmd,
		stderr:   errLines,
		stdout:   outLines,
		combined: output,
	}, nil
}

type Setup struct {
	containerName string
	cleanUpFunc   func()
}

func setup(t *testing.T) *Setup {
	containerName := "runapp-e2e-" + strings.ToLower(t.Name())
	_, err := runCommand(fmt.Sprintf("docker rm -f %s", containerName))
	require.NoError(t, err)

	_, err = runCommand(fmt.Sprintf("docker create --name %s robertdebock/ubuntu sleep infinity", containerName))
	require.NoError(t, err)

	_, err = runCommand(fmt.Sprintf("docker cp ../bin/runapp %s:/usr/local/bin/runapp", containerName))
	require.NoError(t, err)

	_, err = runCommand(fmt.Sprintf("docker start %s", containerName))
	require.NoError(t, err)

	return &Setup{
		containerName: containerName,
		cleanUpFunc: func() {
			// cleanup
			_, err := runCommand(fmt.Sprintf("docker rm -f %s", containerName))
			require.NoError(t, err)
		},
	}
}
