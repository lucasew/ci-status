package reporter

import (
	"fmt"
	"os"
)

// ReportError centralizes error reporting for the application.
// This function fulfills the requirement for a single error-reporting funnel.
// It logs the error to stderr and can be wired to Sentry or other backends later.
func ReportError(err error) {
	if err == nil {
		return
	}
	// In the future, this is where we would call Sentry.captureException(err)
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// ReportWarning centralizes warning reporting.
func ReportWarning(err error, format string, args ...any) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}
