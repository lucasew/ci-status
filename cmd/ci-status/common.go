package main

import (
	"fmt"
	"os"

	"ci-status/internal/forge"
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

// initForge centralizes the logic for detecting and initializing the forge
// client and the commit SHA. It returns a nil client if not in a CI
// environment or if detection fails.
func initForge(forgeOverride, commitOverride string, silent bool) (forge.ForgeClient, string) {
	if !isCI(silent) {
		return nil, ""
	}

	client, err := forge.DetectClient(forgeOverride)
	if err != nil {
		if !silent {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
		return nil, ""
	}

	commit, err := forge.DetectCommit(commitOverride)
	if err != nil {
		if !silent {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
		// Allow returning a client even if commit detection fails,
		// but the commit string will be empty.
	}

	return client, commit
}
