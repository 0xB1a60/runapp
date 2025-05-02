package tui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
)

var (
	ErrStop = errors.New("stop")
)

type FlagOrPromptEntry struct {
	Value string

	TUIFunc      func() (*string, error)
	ValidateFunc func(string) error
	SetFunc      func(value string)
}

func ResolveFlagsOrPrompt(values ...FlagOrPromptEntry) error {
	for _, value := range values {
		if len(value.Value) == 0 {
			val, err := value.TUIFunc()
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					return ErrStop
				}
				if strings.Contains(err.Error(), "could not open a new TTY") {
					if err := value.ValidateFunc(value.Value); err != nil {
						return err
					}
					continue
				}
				return err
			}
			value.SetFunc(*val)
			continue
		}

		if err := value.ValidateFunc(value.Value); err != nil {
			return err
		}
	}
	return nil
}
