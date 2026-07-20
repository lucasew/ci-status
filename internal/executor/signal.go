package executor

import (
	"context"
	"os/signal"
)

// withSignalCancel derives a context that is cancelled when the process receives
// an interrupt-like signal (see interruptSignals). Callers must invoke the
// returned stop function to restore default signal handling.
func withSignalCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, interruptSignals...)
}
