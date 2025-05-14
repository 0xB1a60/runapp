package logs

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/liamg/tml"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/util"
)

func PrintLines(app apps.App) error {
	if err := printFile(os.Stdout, app.StdoutPath, false); err != nil {
		return err
	}
	if err := printFile(os.Stderr, app.StderrPath, true); err != nil {
		return err
	}
	return nil
}

func printFile(w io.Writer, filename string, asError bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			util.DebugLog("Failed to close file %s: %s", filename, err)
		}
	}(file)

	scanner := bufio.NewScanner(file)

	// expand the buffer
	buf := make([]byte, maxBufferCapacity)
	scanner.Buffer(buf, maxBufferCapacity)

	for scanner.Scan() {
		if asError {
			if _, err := fmt.Fprintln(w, tml.Sprintf("<red>%s</red>", scanner.Text())); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintln(w, scanner.Text()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}
	return nil
}
