package main

import (
	"fmt"
	"os"
)

func isCI(silent bool) bool {
	if os.Getenv("CI") == "" {
		if !silent {
			fmt.Fprintln(os.Stderr, "Warning: CI environment variable not set, skipping status reporting")
		}
		return false
	}
	return true
}
