package main

import (
	"os"

	"github.com/0xB1a60/runapp/internal/cli"
)

var version = "main"

func main() {
	if err := cli.Start(version); err != nil {
		os.Exit(1)
	}
}
