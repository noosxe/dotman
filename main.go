package main

import (
	"os"

	"github.com/noosxe/dotman/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
