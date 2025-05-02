package apps

import (
	"context"

	"github.com/nxadm/tail"
)

type Log struct {
	Value string
	IsErr bool
}

func ReadLogs(ctx context.Context, app App) (<-chan Log, error) {
	tailCfg := tail.Config{
		Follow:        true,
		ReOpen:        true,
		MustExist:     false,
		CompleteLines: true,
		Logger:        tail.DiscardingLogger, // ignore logs from tail itself
	}

	outTail, err := tail.TailFile(app.StdoutPath, tailCfg)
	if err != nil {
		return nil, err
	}

	errTail, err := tail.TailFile(app.StderrPath, tailCfg)
	if err != nil {
		return nil, err
	}

	ch := make(chan Log, 100)

	go func() {
		for {
			select {
			case outLine := <-outTail.Lines:
				ch <- Log{Value: outLine.Text, IsErr: false}
			case errLine := <-errTail.Lines:
				ch <- Log{Value: errLine.Text, IsErr: true}
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}
