package main

import (
	"os"

	"github.com/jepemo/miko-shell/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
