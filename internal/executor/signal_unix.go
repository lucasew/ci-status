//go:build unix

package executor

import (
	"os"
	"syscall"
)

// interruptSignals are the signals that should tear down a wrapped command.
// SIGINT covers Ctrl+C; SIGTERM covers container/CI stops; SIGHUP covers
// terminal/session hangup (SSH disconnect). The child runs in its own process
// group (Setpgid), so it does not receive those signals with the parent —
// NotifyContext must cancel so killCommand can still reap the tree.
var interruptSignals = []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGHUP}
