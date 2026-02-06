package errors

import (
	"fmt"
	"log/slog"
)

// Report reports an error to the centralized monitoring system.
// It logs to stderr as a fallback.
func Report(err error) {
	if err == nil {
		return
	}
	// Placeholder for Sentry integration
	// sentry.CaptureException(err)
	slog.Error("error reported", "err", err)
}

// Warn logs a warning message to stderr.
// This is for non-critical errors that don't stop execution.
func Warn(err error) {
	if err == nil {
		return
	}
	slog.Warn("warning reported", "err", err)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	slog.Warn(msg)
}
