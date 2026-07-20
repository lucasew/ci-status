//go:build unix

package executor

import (
	"os"
	"syscall"
)

// interruptSignals are the signals that should tear down a wrapped command.
// SIGINT covers Ctrl+C; SIGTERM covers container/CI stops.
var interruptSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
