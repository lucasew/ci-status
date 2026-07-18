package main

import (
	"errors"
	"fmt"
	"os"

	"ci-status/internal/forge"
)

// quietError marks an error that should still fail the process (non-zero exit)
// but must not be printed. Used when --silent is set so CI scripts can check
// exit codes without stderr noise.
type quietError struct {
	err error
}

func (e quietError) Error() string { return e.err.Error() }
func (e quietError) Unwrap() error { return e.err }

// quiet returns err unchanged, or wraps it so main suppresses printing when silent.
func quiet(err error, silent bool) error {
	if err == nil || !silent {
		return err
	}
	return quietError{err: err}
}

func isQuietError(err error) bool {
	var q quietError
	return errors.As(err, &q)
}

// isCI checks if the tool is running inside a Continuous Integration environment.
// It relies on the presence of the "CI" environment variable (standard in GitHub Actions, GitLab CI, etc.).
// If silent is false, it prints a warning to stderr when CI is not detected.
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
