package forge

import (
	"context"
)

type State string

const (
	StateRunning State = "running"
	StatePending State = "pending"
	StateSuccess State = "success"
	StateFailure State = "failure"
	StateError   State = "error"
)

type StatusOpts struct {
	Commit      string
	Context     string
	State       State
	Description string
	TargetURL   string
}

type ForgeClient interface {
	SetStatus(ctx context.Context, opts StatusOpts) error
}

// ForgeLoader is a strategy function that attempts to create a ForgeClient from a remote URL.
// It returns nil if the URL is not supported by this strategy.
type ForgeLoader func(url string) ForgeClient
