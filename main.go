package main

import (
	"os"

	"github.com/rubrical-studios/gh-pmu/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
