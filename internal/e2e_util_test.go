package internal

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/moby/moby/pkg/stdcopy"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/util"
)

var (
	inContainerAppPath = "/usr/local/bin/runapp"
)

type execResult struct {
	stdout   string
	stderr   string
	exitCode int
}

type SetupResult struct {
	container testcontainers.Container

	exec        func(v string, options ...exec.ProcessOption) execResult
	listApps    func() []apps.App
	cleanUpFunc func()
}

func setup(t *testing.T) *SetupResult {
	req := testcontainers.ContainerRequest{
		Image: "ubuntu:24.04",
		Cmd:   []string{"sleep", "infinity"},
	}
	container, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	fromPath, err := util.ResolvePath("../bin/runapp")
	require.NoError(t, err)

	err = container.CopyFileToContainer(t.Context(), fromPath, inContainerAppPath, 0777)
	require.NoError(t, err)

	execAction := func(arg string, options ...exec.ProcessOption) execResult {
		exitCode, r, err := container.Exec(t.Context(), []string{"/bin/sh", "-c", inContainerAppPath + " " + arg}, options...)
		require.NoError(t, err)

		// Use stdcopy.StdCopy to demultiplex stdout and stderr
		var stdoutBuf, stderrBuf bytes.Buffer
		_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, r)
		require.NoError(t, err)

		return execResult{
			exitCode: exitCode,
			stdout:   strings.TrimSuffix(stdoutBuf.String(), "\n"),
			stderr:   strings.TrimSuffix(stderrBuf.String(), "\n"),
		}
	}

	return &SetupResult{
		container: container,
		exec:      execAction,
		listApps: func() []apps.App {
			execRes := execAction("--json")
			require.Equal(t, 0, execRes.exitCode)

			var list []apps.App
			require.NoError(t, json.Unmarshal([]byte(execRes.stdout), &list))
			return list
		},
		cleanUpFunc: func() {
			require.NoError(t, container.Terminate(t.Context()))
		},
	}
}

func addShellSH() exec.ProcessOption {
	return exec.WithEnv([]string{
		"SHELL=/bin/sh",
	})
}
