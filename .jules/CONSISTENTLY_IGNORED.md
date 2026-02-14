## IGNORE: Path Traversal Fixes via Sanitization

**- Pattern:** Attempting to fix path traversal vulnerabilities in URL parsing by using `path.Clean`, `strings.Replace`, or `strings.Trim` to sanitize input strings.
**- Justification:** Multiple attempts (PRs #17, #18, #19, #26) using these methods were rejected. Sanitization is error-prone and often misses edge cases. The preferred approach (as seen in PR #29) is **Strict Allowlisting** (e.g., using Regex `^[a-zA-Z0-9_.-]+$`) to reject any input that does not conform to the expected format.
**- Files Affected:** `internal/forge/*.go`

## IGNORE: Redundant Docstrings

**- Pattern:** Adding documentation comments that simply restate the function name or describe "what" the code does in plain English (e.g., `// execute runs the command`).
**- Justification:** PR #23 was rejected in favor of PR #25. Documentation must provide context, explaining the **"Why"** (design decisions), **"Flow"** (high-level logic), and **"Side Effects"**. Comments that do not add understanding are considered noise.
**- Files Affected:** `*.go`

## IGNORE: "Clean Code" Refactoring of CLI Logic

**- Pattern:** Refactoring linear procedural code (splitting into helpers, abstracting `os.Exit`) solely for architectural purity or testability.
**- Justification:** PRs #15 and #20 were rejected. The project prioritizes **Locality of Behavior**. Logic should remain linear and self-contained within the command. Fragmenting flow or abstracting exit logic forces the reader to jump around to understand the sequential execution.
**- Files Affected:** `cmd/ci-status/*.go`

## IGNORE: Misapplied Shared Logic

**- Pattern:** Replacing explicit, context-specific logic in commands (like `set.go`) with generic shared helpers (like `initForge`) that introduce invalid constraints.
**- Justification:** PR #24 was rejected. While DRY is generally good, helpers like `initForge` often enforce constraints (e.g., `isCI` checks) that are not valid for all contexts (e.g., `set` command supports manual/local execution).
**- Files Affected:** `cmd/ci-status/set.go`, `cmd/ci-status/common.go`

## IGNORE: Toolchain Version Changes

**- Pattern:** Modifying the Go version in `go.mod` or `mise.toml` to a different version (e.g., downgrading or changing patch versions) without explicit instruction.
**- Justification:** PR #10 modified the Go version and was closed. The project relies on specific toolchain versions pinned in `mise.toml` to ensure consistent builds across environments. Arbitrary version changes cause environment drift.
**- Files Affected:** `go.mod`, `mise.toml`
