package forge

import (
	"context"
)

// State represents the status of a CI/CD pipeline step reported to the forge.
// Different forges might map these states slightly differently (e.g. GitHub treats 'running' as 'pending').
type State string

const (
	// StateRunning indicates the task is currently executing.
	// Note: GitHub API maps this to 'pending' with a description unless using check runs.
	StateRunning State = "running"
	// StatePending indicates the task is queued or waiting.
	StatePending State = "pending"
	// StateSuccess indicates the task completed successfully (exit code 0).
	StateSuccess State = "success"
	// StateFailure indicates the task failed (non-zero exit code).
	StateFailure State = "failure"
	// StateError indicates a configuration or runtime error prevented the task from running properly.
	StateError   State = "error"
)

// StatusOpts encapsulates the parameters required to set a commit status.
type StatusOpts struct {
	// Commit is the SHA-1 hash of the commit to update.
	Commit      string
	// Context is the label that differentiates this status check (e.g., "ci/lint").
	Context     string
	// State is the current status of the task.
	State       State
	// Description is a short, human-readable summary of the status.
	Description string
	// TargetURL is an optional link to the build details (e.g., CI logs).
	TargetURL   string
}

// ForgeClient defines the interface for interacting with a Git forge (GitHub, GitLab, Gitea, etc.).
// Implementations handle the specifics of authentication and API calls.
type ForgeClient interface {
	// SetStatus updates the commit status for the given options.
	// It should handle API-specific nuances, such as state mapping.
	SetStatus(ctx context.Context, opts StatusOpts) error
}

// ForgeLoader is a strategy function that attempts to instantiate a ForgeClient from a remote URL.
// It returns nil if the URL is not supported by this strategy, allowing the next strategy to be tried.
type ForgeLoader func(url string) ForgeClient
