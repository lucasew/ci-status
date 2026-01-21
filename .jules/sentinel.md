## 2024-09-05 - Inadequate Sanitization via TrimSpace

**Vulnerability:** HTTP Header Injection (CRLF Injection) in the GitHub client's `Authorization` header construction.
**Learning:** An attempt to simplify the sanitization logic by replacing `strings.NewReplacer("\n", "", "\r", "")` with `strings.TrimSpace()` introduced a critical security flaw. `strings.TrimSpace()` only removes leading and trailing whitespace, including `\r` and `\n`. It does *not* remove these characters if they are present in the middle of the token, leaving the application vulnerable to header injection.
**Prevention:** When sanitizing inputs to prevent CRLF injection, always use a method that removes newline and carriage return characters from the *entire* string, not just the beginning and end. `strings.NewReplacer` is a robust and appropriate tool for this purpose. Do not assume that `strings.TrimSpace` provides equivalent security. Always verify that sanitization logic correctly addresses the specific threat model.

## 2026-01-16 - Path Traversal in Generic Remote Parsing

**Vulnerability:** Path Traversal in `ParseGenericRemote` logic.
**Learning:** The initial implementation of `ParseGenericRemote` relied on manual string splitting (`strings.FieldsFunc`) without resolving path traversal sequences (`..`). This allowed crafted URLs (e.g., `https://example.com/owner/repo/../../attacker/project`) to manipulate the detected owner and repository values, potentially causing the application to interact with unexpected endpoints or contexts.
**Prevention:** Always use standard library functions like `net/url.Parse` and `path.Clean` to normalize and sanitize URLs before extracting path components. Avoid manual string manipulation for path parsing as it is error-prone and often misses edge cases like `..` traversal.
