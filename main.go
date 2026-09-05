package main

import (
	"os"

	"github.com/Malay-dev/gitx-profile/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
