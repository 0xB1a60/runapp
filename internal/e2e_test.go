package internal

import (
	"context"
	"testing"

	"github.com/0xB1a60/runapp/internal/common"
	"github.com/stretchr/testify/require"
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
		"yaml": {
			extraParams: "--yaml",
			expected:    "[]\n",
		},
		"json": {
			extraParams: "--json",
			expected:    "[]",
		},
	}

	for name, tCase := range cases {
		t.Run(name, func(t *testing.T) {

			res := s.exec(tCase.extraParams)
			require.Equal(t, tCase.expected, res.stdout)
			require.Empty(t, res.stderr)
			require.Equal(t, 0, res.exitCode)
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
	}

	runRes := s.exec(`run fapp --start-on-boot --command 'exit 1;'`, addShellSH())
	require.Equal(t, 0, runRes.exitCode)

	cases := map[string]testCase{
		"default": {
			expected: "Error: app: notexistapp does not exist",
		},
		"json": {
			extraParams: "--json",
			expected:    `{"error":"not_found"}`,
		},
		"yaml": {
			extraParams: "--yaml",
			expected:    "error: not_found\n",
		},
	}

	for name, tCase := range cases {
		t.Run(name, func(t *testing.T) {

			res := s.exec("status notexistapp " + tCase.extraParams)
			require.Equal(t, tCase.expected, res.stderr)
			require.Equal(t, 1, res.exitCode)
		})
	}
}

func TestNonExistentApp(t *testing.T) {
	t.Parallel()

	s := setup(t)
	defer s.cleanUpFunc()

	runRes := s.exec(`run fapp --start-on-boot --command 'exit 1;'`, addShellSH())
	require.Equal(t, 0, runRes.exitCode)

	type testCase struct {
	}

	cases := map[string]testCase{
		"kill":   {},
		"logs":   {},
		"remove": {},
	}

	for name := range cases {
		t.Run(name, func(t *testing.T) {
			res := s.exec(name + " notexistapp")
			require.Equal(t, 0, runRes.exitCode)
			require.Equal(t, "Error: app: notexistapp does not exist", res.stderr)
		})
	}
}

func TestLogs(t *testing.T) {
	t.Parallel()

	s := setup(t)
	defer s.cleanUpFunc()

	runRes := s.exec(`run my-app --start-on-boot --command 'echo "stdout"; echo "stderr" >&2'`, addShellSH())
	require.Equal(t, 0, runRes.exitCode)

	res := s.exec(" logs my-app")
	require.Contains(t, res.stdout, "stdout")
	require.Contains(t, res.stderr, "\x1b[0m\x1b[31mstderr\x1b[39m\x1b[0m")
}

func TestRemoveMany(t *testing.T) {
	t.Parallel()

	s := setup(t)
	defer s.cleanUpFunc()

	appRes := s.exec(`run sapp --start-on-boot --command 'exit 0;'`, addShellSH())
	require.Equal(t, 0, appRes.exitCode)

	appRes = s.exec(`run fapp --start-on-boot --command 'exit 1;'`, addShellSH())
	require.Equal(t, 0, appRes.exitCode)

	appRes = s.exec(`run rapp --skip-logs --start-on-boot --command 'sleep 1200'`, addShellSH())
	require.Equal(t, 0, appRes.exitCode)

	list := s.listApps()
	require.Len(t, list, 3)

	require.Contains(t, list[0].Name, "rapp")
	require.Contains(t, list[1].Name, "fapp")
	require.Contains(t, list[2].Name, "sapp")

	removeManyFailedRes := s.exec(`removemany --failed`)
	require.Equal(t, 0, removeManyFailedRes.exitCode)

	list = s.listApps()
	require.Len(t, list, 2)
	require.Contains(t, list[0].Name, "rapp")
	require.Contains(t, list[1].Name, "sapp")

	removeManySuccessRes := s.exec(`removemany --success`)
	require.Equal(t, 0, removeManySuccessRes.exitCode)

	list = s.listApps()
	require.Len(t, list, 1)
	require.Contains(t, list[0].Name, "rapp")
}

func TestFlow(t *testing.T) {
	t.Parallel()

	s := setup(t)
	defer s.cleanUpFunc()

	appRes := s.exec(`run my-app --skip-logs --start-on-boot --command 'sleep 1200'`, addShellSH())
	require.Equal(t, 0, appRes.exitCode)

	// get current status
	list := s.listApps()
	require.Len(t, list, 1)
	require.Equal(t, "my-app", list[0].Name)
	require.Equal(t, common.AppStatusRunning, list[0].Status)

	// kill it
	killRes := s.exec(`kill my-app`, addShellSH())
	require.Equal(t, 0, killRes.exitCode)

	// get status after kill
	list = s.listApps()
	require.Len(t, list, 1)
	require.Equal(t, "my-app", list[0].Name)
	require.Equal(t, common.AppStatusFailed, list[0].Status)

	// restart it
	restartRes := s.exec(`restart my-app --skip-logs`, addShellSH())
	require.Equal(t, 0, restartRes.exitCode)

	// get status after restart
	list = s.listApps()
	require.Len(t, list, 1)
	require.Equal(t, "my-app", list[0].Name)
	require.Equal(t, common.AppStatusRunning, list[0].Status)

	// simulate onboot
	require.NoError(t, s.container.Stop(context.Background(), nil))
	require.NoError(t, s.container.Start(context.Background()))

	onBootRes := s.exec(`onboot`, addShellSH())
	require.Equal(t, 0, onBootRes.exitCode)

	// get status after restart
	list = s.listApps()
	require.Len(t, list, 1)
	require.Equal(t, "my-app", list[0].Name)
	require.Equal(t, common.AppStatusRunning, list[0].Status)

	// kill it again
	killRes = s.exec(`kill my-app`, addShellSH())
	require.Equal(t, 0, killRes.exitCode)

	// remove it
	killRes = s.exec(`remove my-app`, addShellSH())
	require.Equal(t, 0, killRes.exitCode)

	// get status after removal
	list = s.listApps()
	require.Len(t, list, 0)
}
