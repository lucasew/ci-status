package errors

import (
	"fmt"
	"io"
	"os"
)

// Writer is the destination for error logs. Defaults to os.Stderr.
var Writer io.Writer = os.Stderr

// Report reports an error to the centralized monitoring system.
// It logs to stderr as a fallback.
func Report(err error) {
	if err == nil {
		return
	}
	// Placeholder for Sentry integration
	// sentry.CaptureException(err)
	fmt.Fprintf(Writer, "Error: %v\n", err)
}

// Warn logs a warning message to stderr.
// This is for non-critical errors that don't stop execution.
func Warn(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(Writer, "Warning: %v\n", err)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	fmt.Fprintf(Writer, "Warning: "+format+"\n", args...)
}
