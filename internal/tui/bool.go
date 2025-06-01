package tui

import (
	"github.com/charmbracelet/huh"
)

const (
	Yes = "yes"
	No  = "no"
)

func OnBool(title string) (bool, error) {
	var value bool
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[bool]().
			Title(title).
			Options(
				huh.NewOption(No, false),
				huh.NewOption(Yes, true),
			).
			Value(&value),
	)).WithTheme(huh.ThemeBase())
	if err := form.Run(); err != nil {
		return false, err
	}
	return value, nil
}
