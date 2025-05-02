package tui

import (
	"github.com/charmbracelet/huh"

	"github.com/0xB1a60/runapp/internal/apps"
)

func NamePicker() (*string, error) {
	list, err := apps.List()
	if err != nil {
		return nil, err
	}

	options := make([]huh.Option[string], 0, len(list))
	for _, app := range list {
		options = append(options, huh.NewOption(app.Name, app.Name))
	}

	var value string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("My app is:").
			Options(options...).
			Value(&value),
	)).WithTheme(huh.ThemeBase())
	if err := form.Run(); err != nil {
		return nil, err
	}
	return &value, nil
}
