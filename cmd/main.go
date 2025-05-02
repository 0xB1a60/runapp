package main

import (
	"os"

	"github.com/0xB1a60/runapp/internal/cli"
)

func main() {
	if err := cli.Start(); err != nil {
		os.Exit(1)
	}
}
