package main

import (
	"os"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
