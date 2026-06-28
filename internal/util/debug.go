package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/liamg/tml"
)

const (
	debugEnv = "DEBUG"
)

var (
	cachedDebug *bool
)

func IsDebug() bool {
	if cachedDebug != nil {
		return *cachedDebug
	}

	val, ok := os.LookupEnv(debugEnv)
	if !ok {
		cachedDebug = new(false)
	} else {
		cachedDebug = new(strings.EqualFold(val, "true"))
	}
	return *cachedDebug
}

func DebugLog(format string, args ...any) {
	if IsDebug() {
		fmt.Println(tml.Sprintf("<purple>%s</purple>", fmt.Sprintf(format, args...)))
	}
}
