package config

import (
	"time"
)

// Config holds the configuration for the 'run' command execution.
type Config struct {
	// ContextName is the status context label (e.g. "ci/test").
	ContextName string
	// Command is the binary to execute.
	Command     string
	// Args are arguments passed to the command.
	Args        []string

	// Forge overrides the detected forge (e.g. "github").
	Forge       string
	// Commit overrides the detected commit SHA.
	Commit      string
	// PR overrides the pull request number (optional).
	PR          string
	// URL is the target URL for the status details.
	URL         string
	// PendingDesc is the description set when the command starts.
	PendingDesc string
	// SuccessDesc is the description set when the command exits with 0.
	SuccessDesc string
	// FailureDesc is the description set when the command exits with non-zero.
	FailureDesc string
	// Timeout is the maximum duration for the command execution.
	Timeout     time.Duration
	// Silent suppresses non-critical output (warnings).
	Silent      bool
}
