package logs

import (
	"context"

	"github.com/nxadm/tail"

	"github.com/0xB1a60/runapp/internal/apps"
)

type Log struct {
	Value string
	IsErr bool
}

// Stream reads the logs from the given app's stdout and stderr files
func Stream(ctx context.Context, app apps.App, logType LogType) (<-chan Log, error) {
	tailCfg := tail.Config{
		Follow:        true,
		ReOpen:        true,
		MustExist:     false,
		CompleteLines: true,
		MaxLineSize:   maxBufferCapacity,
		Logger:        tail.DiscardingLogger, // ignore logs from tail itself
	}

	ch := make(chan Log, 100)

	if logType == OutLogs {
		outTail, err := tail.TailFile(app.StdoutPath, tailCfg)
		if err != nil {
			return nil, err
		}

		go func() {
			for {
				select {
				case outLine := <-outTail.Lines:
					if outLine == nil {
						return
					}
					ch <- Log{Value: outLine.Text, IsErr: false}
				case <-ctx.Done():
					return
				}
			}
		}()
		return ch, nil
	}

	if logType == ErrLogs {
		errTail, err := tail.TailFile(app.StderrPath, tailCfg)
		if err != nil {
			return nil, err
		}

		go func() {
			for {
				select {
				case errLine := <-errTail.Lines:
					if errLine == nil {
						return
					}
					ch <- Log{Value: errLine.Text, IsErr: true}
				case <-ctx.Done():
					return
				}
			}
		}()
		return ch, nil
	}

	outTail, err := tail.TailFile(app.StdoutPath, tailCfg)
	if err != nil {
		return nil, err
	}

	errTail, err := tail.TailFile(app.StderrPath, tailCfg)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case outLine := <-outTail.Lines:
				if outLine == nil {
					return
				}
				ch <- Log{Value: outLine.Text, IsErr: false}
			case errLine := <-errTail.Lines:
				if errLine == nil {
					return
				}
				ch <- Log{Value: errLine.Text, IsErr: true}
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}
