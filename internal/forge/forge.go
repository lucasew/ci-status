package forge

import (
	"context"
)

type State string

const (
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
