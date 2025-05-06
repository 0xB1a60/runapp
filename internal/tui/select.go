package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/liamg/tml"

	"github.com/0xB1a60/runapp/internal/apps"
	"github.com/0xB1a60/runapp/internal/common"
)

func NamePicker() (*string, error) {
	return NamePickerWithValidator(nil)
}

func NamePickerWithValidator(validate func([]apps.App, string) error) (*string, error) {
	list, err := apps.List()
	if err != nil {
		return nil, err
	}

	options := make([]huh.Option[string], 0, len(list))
	for _, app := range list {
		var key string
		switch app.Status {
		case common.AppStatusFailed:
			key = tml.Sprintf("<red>%s</red>", app.Name)
		case common.AppStatusSuccess:
			key = tml.Sprintf("<green>%s</green>", app.Name)
		case common.AppStatusStarting:
			key = tml.Sprintf("<yellow>%s</yellow>", app.Name)
		case common.AppStatusRunning:
			key = tml.Sprintf("<yellow>%s</yellow>", app.Name)
		}
		options = append(options, huh.NewOption(key, app.Name))
	}

	var value string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("My app is:").
			Options(options...).
			Validate(func(s string) error {
				if validate == nil {
					return nil
				}
				return validate(list, s)
			}).
			Value(&value),
	)).WithTheme(huh.ThemeBase())
	if err := form.Run(); err != nil {
		return nil, err
	}
	return &value, nil
}
