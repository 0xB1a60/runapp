package util

import (
	"os"
	"strings"
)

var posixShells = map[string]struct{}{
	"/usr/bin/sh": {},
	"/bin/sh":     {},

	"/usr/bin/ash": {},
	"/bin/ash":     {},

	"/usr/bin/bash": {},
	"/bin/bash":     {},

	"/usr/bin/dash": {},
	"/bin/dash":     {},

	"/usr/bin/zsh": {},
	"/bin/zsh":     {},

	"/usr/bin/ksh": {},
	"/bin/ksh":     {},

	"/usr/bin/fish": {},
	"/bin/fish":     {},

	"/usr/bin/tcsh": {},
	"/bin/tcsh":     {},

	"/usr/bin/csh": {},
	"/bin/csh":     {},
}

func GetShellArgs() []string {
	for _, env := range os.Environ() {
		parts := strings.Split(env, "=")
		if len(parts) > 1 && parts[0] == "SHELL" {
			if _, ok := posixShells[parts[1]]; ok {
				return []string{parts[1], "-c"}
			}
			break
		}
	}
	return nil
}
