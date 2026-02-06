package main

import (
	"os"

	reporter "ci-status/internal/errors"
	"ci-status/internal/forge"
)

// isCI checks if the tool is running inside a Continuous Integration environment.
// It relies on the presence of the "CI" environment variable (standard in GitHub Actions, GitLab CI, etc.).
// If silent is false, it prints a warning to stderr when CI is not detected.
func isCI(silent bool) bool {
	if os.Getenv("CI") == "" {
		if !silent {
			reporter.Warnf("CI environment variable not set, skipping status reporting")
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
			reporter.Warn(err)
		}
		return nil, ""
	}

	commit, err := forge.DetectCommit(commitOverride)
	if err != nil {
		if !silent {
			reporter.Warn(err)
		}
		// Allow returning a client even if commit detection fails,
		// but the commit string will be empty.
	}

	return client, commit
}
