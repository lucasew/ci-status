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

// initForge handles the common initialization logic for forge interactions.
//
// It performs the following steps:
// 1. Checks if running in a CI environment (unless silent).
// 2. Detects the appropriate ForgeClient.
// 3. Detects the target commit SHA.
//
// It returns the client and commit SHA. If any step fails (and isn't fatal),
// it may return a nil client or empty commit, logging warnings unless silent.
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
