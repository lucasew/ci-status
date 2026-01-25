package forge

import (
	"context"
)

// State represents the status of a CI/CD pipeline stage.
type State string

const (
	// StateRunning indicates the task is currently executing.
	// Note: GitHub API treats "running" as "pending", but some forges might distinguish them.
	StateRunning State = "running"
	// StatePending indicates the task is waiting to start or in progress.
	StatePending State = "pending"
	// StateSuccess indicates the task completed successfully (exit code 0).
	StateSuccess State = "success"
	// StateFailure indicates the task failed (non-zero exit code).
	StateFailure State = "failure"
	// StateError indicates an internal error occurred (e.g. timeout, configuration error).
	StateError   State = "error"
)

// StatusOpts contains the parameters for setting a commit status.
type StatusOpts struct {
	// Commit is the full SHA of the commit to update.
	Commit      string
	// Context is the label shown in the UI (e.g. "ci/lint").
	Context     string
	// State is the status to report.
	State       State
	// Description is a short human-readable message.
	Description string
	// TargetURL is an optional link to more details (e.g. build logs).
	TargetURL   string
}

// ForgeClient is the abstraction for communicating with Git forges (GitHub, GitLab, Gitea, etc.).
type ForgeClient interface {
	// SetStatus updates the status of a specific commit.
	SetStatus(ctx context.Context, opts StatusOpts) error
}

// ForgeLoader is a strategy function that attempts to create a ForgeClient from a remote URL.
// It returns nil if the URL is not supported by this strategy.
type ForgeLoader func(url string) ForgeClient
