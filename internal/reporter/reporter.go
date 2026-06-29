package reporter

import (
	"fmt"
	"log/slog"
)

// ReportError centralizes error reporting for the application.
// This function fulfills the requirement for a single error-reporting funnel.
// It logs the error to stderr and can be wired to Sentry or other backends later.
func ReportError(err error) {
	if err == nil {
		return
	}
	// In the future, this is where we would call Sentry.captureException(err)
	slog.Error(err.Error())
}

// ReportWarning centralizes warning reporting.
func ReportWarning(err error, format string, args ...any) {
	if err == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	slog.Warn(msg)
}
