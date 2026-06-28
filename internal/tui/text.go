package tui

import (
	"errors"

	"github.com/charmbracelet/huh"
)

func TextWithPlaceholder(title string, placeholder string) func() (*string, error) {
	return func() (*string, error) {
		var value string
		form := huh.NewForm(huh.NewGroup(
			huh.NewText().
				Title(title).
				Lines(1).
				Placeholder(placeholder).
				Validate(func(s string) error {
					if len(s) == 0 && len(placeholder) == 0 {
						return errors.New("please enter value")
					}
					return nil
				}).
				Value(&value),
		)).WithTheme(huh.ThemeBase())
		if err := form.Run(); err != nil {
			return nil, err
		}

		if len(value) == 0 {
			return new(placeholder), nil
		}
		return &value, nil
	}
}
