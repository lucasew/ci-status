# Agent Instructions

## Error Handling
- **No Silent Failures:** Never ignore errors. All errors must be handled or reported.
- **Centralized Reporting:** Use `ci-status/internal/errors` (aliased as `reporter`) for all error reporting.
    - Use `reporter.Report(err)` for errors.
    - Use `reporter.Warn(err)` or `reporter.Warnf(...)` for warnings.
    - Do not use `fmt.Fprintf(os.Stderr, ...)` directly for errors/warnings.
