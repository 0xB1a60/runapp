package internal

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
)

func TestListEmpty(t *testing.T) {
	t.Parallel()

	s := setup(t)
	defer s.cleanUpFunc()

	type testCase struct {
		extraParams string
		expected    string
	}

	cases := map[string]testCase{
		"default": {
			expected: common.NoAppsMessage,
		},
		"json": {
			extraParams: "--json",
			expected:    "[]",
		},
		"yaml": {
			extraParams: "--yaml",
			expected:    "[]",
		},
	}

	for name, tCase := range cases {
		t.Run(name, func(t *testing.T) {

			commandRes, err := runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp "+tCase.extraParams, s.containerName))
			require.NoError(t, err)
			require.NotEmpty(t, commandRes.combined)
			require.Equal(t, tCase.expected, commandRes.combined[0])
		})
	}
}

func TestStatus_NonExistent(t *testing.T) {
	t.Parallel()

	s := setup(t)
	defer s.cleanUpFunc()

	type testCase struct {
		extraParams string
		expected    string
		expectedErr bool
	}

	cases := map[string]testCase{
		"default": {
			expected: common.NoAppsMessage,
		},
		"json": {
			extraParams: "--json",
			expectedErr: true,
		},
		"yaml": {
			extraParams: "--yaml",
			expectedErr: true,
		},
	}

	for name, tCase := range cases {
		t.Run(name, func(t *testing.T) {

			commandRes, err := runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp status --name notexistapp "+tCase.extraParams, s.containerName))
			if tCase.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, commandRes.combined)
				require.Equal(t, tCase.expected, commandRes.combined[0])
			}
		})
	}
}

func TestNonExistentApp(t *testing.T) {
	t.Parallel()

	s := setup(t)
	defer s.cleanUpFunc()

	type testCase struct {
	}

	cases := map[string]testCase{
		"kill":   {},
		"logs":   {},
		"remove": {},
	}

	for name := range cases {
		t.Run(name, func(t *testing.T) {

			_, err := runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp %s --name notexistapp", name, s.containerName))
			require.Error(t, err)
		})
	}

}

func TestLogs(t *testing.T) {
	t.Parallel()

	s := setup(t)
	defer s.cleanUpFunc()

	_, err := runCommand(fmt.Sprintf(`docker exec %s /bin/bash -c "export SHELL=/bin/bash && runapp run --name my-app --start-on-boot --command 'echo "stdout"; echo "stderr" >&2'"`, s.containerName))
	require.NoError(t, err)

	listRes, err := runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp logs --name my-app", s.containerName))
	require.NoError(t, err)
	require.Contains(t, listRes.stdout, "stdout")
	require.Contains(t, listRes.stderr, "\x1b[0m\x1b[31mstderr\x1b[39m\x1b[0m")
}

func TestFlow(t *testing.T) {
	t.Parallel()

	s := setup(t)
	defer s.cleanUpFunc()

	_, err := runCommand(fmt.Sprintf(`docker exec %s /bin/bash -c "export SHELL=/bin/bash && runapp run --name my-app --skip-logs --start-on-boot --command 'sleep 1200'"`, s.containerName))
	require.NoError(t, err)

	// get current status
	listRes, err := runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp --json", s.containerName))
	require.NoError(t, err)
	require.NotEmpty(t, listRes.combined)

	var list []apps.App
	require.NoError(t, json.Unmarshal([]byte(listRes.combined[0]), &list))
	require.Len(t, list, 1)
	require.Equal(t, "my-app", list[0].Name)
	require.Equal(t, common.AppStatusRunning, list[0].Status)

	// kill it
	_, err = runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp kill --name=my-app", s.containerName))
	require.NoError(t, err)

	// get status after kill
	listRes, err = runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp --json", s.containerName))
	require.NoError(t, err)
	require.NotEmpty(t, listRes.combined)

	require.NoError(t, json.Unmarshal([]byte(listRes.combined[0]), &list))
	require.Len(t, list, 1)
	require.Equal(t, "my-app", list[0].Name)
	require.Equal(t, common.AppStatusFailed, list[0].Status)

	_, err = runCommand(fmt.Sprintf("docker restart %s", s.containerName))
	require.NoError(t, err)

	// simulate onboot
	_, err = runCommand(fmt.Sprintf(`docker exec %s /bin/bash -c "export SHELL=/bin/bash && /usr/local/bin/runapp onboot"`, s.containerName))
	require.NoError(t, err)

	// get status after restart
	listRes, err = runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp --json", s.containerName))
	require.NoError(t, err)
	require.NotEmpty(t, listRes.combined)

	require.NoError(t, json.Unmarshal([]byte(listRes.combined[0]), &list))
	require.Len(t, list, 1)
	require.Equal(t, "my-app", list[0].Name)
	require.Equal(t, common.AppStatusRunning, list[0].Status)

	// kill it again
	_, err = runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp kill --name=my-app", s.containerName))
	require.NoError(t, err)

	// remove it
	_, err = runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp remove --name=my-app", s.containerName))
	require.NoError(t, err)

	// get status after removal
	listRes, err = runCommand(fmt.Sprintf("docker exec %s /usr/local/bin/runapp --json", s.containerName))
	require.NoError(t, err)
	require.NotEmpty(t, listRes.combined)

	require.NoError(t, json.Unmarshal([]byte(listRes.combined[0]), &list))
	require.Len(t, list, 0)
}
