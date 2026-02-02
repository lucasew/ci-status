package config

import (
	"time"
)

// Config holds the runtime configuration for the 'run' command.
// It maps command-line flags and arguments to execution parameters,
// controlling how the status is reported and which command is executed.
type Config struct {
	// ContextName is the status identifier (e.g., "lint", "unit-tests").
	// This appears in the pull request status checks and must be unique per workflow step.
	ContextName string
	// Command is the executable to run (e.g., "go", "npm").
	Command string
	// Args contains arguments for the command.
	Args []string

	// Forge overrides the automatic forge detection strategy (e.g., "github").
	// If set, it bypasses the detection logic in DetectClient.
	Forge string
	// Commit overrides the automatic commit SHA detection.
	// Useful when running in non-standard CI environments where env vars aren't reliable.
	Commit string
	// PR overrides the automatic Pull Request detection (if supported).
	PR string
	// URL provides a target URL (e.g., to build logs) for the status 'Details' link.
	URL string
	// PendingDesc is the description shown while the command is executing.
	PendingDesc string
	// SuccessDesc is the description shown when the command exits with code 0.
	SuccessDesc string
	// FailureDesc is the description shown when the command fails (non-zero exit code).
	FailureDesc string
	// Timeout is the maximum duration allowed for the command execution.
	// If exceeded, the command context is cancelled, the process is terminated,
	// and a 'StateError' status is reported to the forge.
	Timeout time.Duration
	// Silent suppresses non-essential output to stdout/stderr.
	// Errors will still be printed, but warnings (like missing CI env) might be suppressed.
	Silent bool
}
