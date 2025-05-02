package util

import (
	"bufio"
	"io"
	"os"
	"os/exec"
)

func ExecuteCommand(command string, logToDebug bool) error {
	args := append(GetShellArgs(), command)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()

	if logToDebug {
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return err
		}

		readFromPipeFunc := func(r io.Reader) {
			reader := bufio.NewReader(r)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					DebugLog("error reading stdout: %v", err)
					break
				}
				DebugLog("stdout: %s", line)
			}
		}

		go readFromPipeFunc(stdoutPipe)
		go readFromPipeFunc(stderrPipe)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}
