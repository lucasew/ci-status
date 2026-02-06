## 2024-09-05 - Inadequate Sanitization via TrimSpace

**Vulnerability:** HTTP Header Injection (CRLF Injection) in the GitHub client's `Authorization` header construction.
**Learning:** An attempt to simplify the sanitization logic by replacing `strings.NewReplacer("\n", "", "\r", "")` with `strings.TrimSpace()` introduced a critical security flaw. `strings.TrimSpace()` only removes leading and trailing whitespace, including `\r` and `\n`. It does *not* remove these characters if they are present in the middle of the token, leaving the application vulnerable to header injection.
**Prevention:** When sanitizing inputs to prevent CRLF injection, always use a method that removes newline and carriage return characters from the *entire* string, not just the beginning and end. `strings.NewReplacer` is a robust and appropriate tool for this purpose. Do not assume that `strings.TrimSpace` provides equivalent security. Always verify that sanitization logic correctly addresses the specific threat model.

## 2026-02-06 - Silent Failures and Decentralized Error Reporting

**Vulnerability:** Ignored errors and lack of centralized reporting mask critical failures, hindering incident response and observability.
**Learning:** Several code paths in `cmd/ci-status` and `internal/forge` were silently ignoring errors (e.g., `_ = client.SetStatus`) or logging them inconsistently to `stderr`. This prevents automated monitoring from capturing issues, leaving the system in an unknown state during failures.
**Prevention:**
1.  **Unified Reporting:** Implemented `internal/errors` (aliased as `reporter`) to centralize error handling. All errors are now funneled through `reporter.Report` or `reporter.Warn`.
2.  **Explicit Handling:** Replaced ignored error assignments (`_ = ...`) with explicit checks and reporting.
3.  **Defer Safety:** Added error checking to `defer` statements (e.g., `resp.Body.Close()`) to catch resource leaks or commit failures that would otherwise go unnoticed.
