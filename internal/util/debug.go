package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/gotidy/ptr"
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
		cachedDebug = ptr.Bool(false)
	} else {
		cachedDebug = ptr.Of(strings.EqualFold(val, "true"))
	}
	return *cachedDebug
}

func DebugLog(format string, args ...interface{}) {
	if IsDebug() {
		fmt.Println(tml.Sprintf("<purple>%s</purple>", fmt.Sprintf(format, args...)))
	}
}
