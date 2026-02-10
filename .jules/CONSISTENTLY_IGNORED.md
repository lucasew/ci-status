## IGNORE: Path Traversal Fixes via Sanitization

**- Pattern:** Attempting to fix path traversal vulnerabilities in URL parsing by using `path.Clean`, `strings.Replace`, or `strings.Trim` to sanitize input strings.
**- Justification:** Multiple attempts (PRs #17, #18, #19, #26) using these methods were rejected. Sanitization is error-prone and often misses edge cases. The preferred approach (as seen in PR #29) is **Strict Allowlisting** (e.g., using Regex `^[a-zA-Z0-9_.-]+$`) to reject any input that does not conform to the expected format.
**- Files Affected:** `internal/forge/*.go`

## IGNORE: Redundant Docstrings

**- Pattern:** Adding documentation comments that simply restate the function name or describe "what" the code does in plain English (e.g., `// execute runs the command`).
**- Justification:** PR #23 was rejected in favor of PR #25. Documentation must provide context, explaining the **"Why"** (design decisions), **"Flow"** (high-level logic), and **"Side Effects"**. Comments that do not add understanding are considered noise.
**- Files Affected:** `*.go`

## IGNORE: Fragmentation of CLI Logic

**- Pattern:** Refactoring linear procedural code in CLI commands (like `run.go` or `set.go`) into multiple small, single-use helper functions.
**- Justification:** PRs #15 and #24 were closed. The project prefers **Locality of Behavior** for CLI command implementations. Fragmenting the logic forces the reader to jump between functions to understand the sequential flow. Helper functions should only be extracted when there is genuine reuse.
**- Files Affected:** `cmd/ci-status/*.go`

## IGNORE: Toolchain Version Changes

**- Pattern:** Modifying the Go version in `go.mod` or `mise.toml` to a different version (e.g., downgrading or changing patch versions) without explicit instruction.
**- Justification:** PR #10 modified the Go version and was closed. The project relies on specific toolchain versions pinned in `mise.toml` to ensure consistent builds across environments. Arbitrary version changes cause environment drift.
**- Files Affected:** `go.mod`, `mise.toml`
