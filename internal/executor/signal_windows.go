//go:build windows

package executor

import "os"

// interruptSignals: Windows delivers Ctrl+C as os.Interrupt.
var interruptSignals = []os.Signal{os.Interrupt}
