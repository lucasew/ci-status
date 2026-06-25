## IGNORE: Path Traversal False Positives in Git URLs

**- Pattern:** Attempting to fix "path traversal" or "URL parsing" vulnerabilities in Git remote URL parsing using sanitization, string manipulation, or even strict allowlisting.
**- Justification:** PRs #17, #18, #19, #26, and #29 were rejected. These are false positive vulnerabilities. The tool parses local git configuration (`git remote get-url origin`), which is considered a trusted source in this context. Do not attempt to add sanitization or strict allowlisting for path traversal here.
**- Files Affected:** `internal/forge/*.go`

## IGNORE: Unwanted Refactoring of CLI Flow

**- Pattern:** Refactoring linear CLI command logic (like `execute` or `initForge`) into smaller helpers, abstracting exit codes, or deduplicating code for "clean code" purity.
**- Justification:** PRs #15, #20, and #24 were rejected. The project prioritizes **Locality of Behavior** over DRY. Logic should remain linear and self-contained within the command. Fragmenting flow makes it harder to read, and shared helpers often introduce invalid constraints (e.g., breaking manual execution in `set.go`).
**- Files Affected:** `cmd/ci-status/*.go`

## IGNORE: Redundant Docstrings

**- Pattern:** Adding documentation comments that simply restate the function name or describe "what" the code does in plain English.
**- Justification:** PR #23 was rejected. Documentation must provide context, explaining the **"Why"** (design decisions), **"Flow"** (high-level logic), and **"Side Effects"**. Comments that do not add understanding are considered noise.
**- Files Affected:** `*.go`

## IGNORE: Toolchain Version Changes

**- Pattern:** Modifying the Go version in `go.mod` or `mise.toml` without explicit instruction.
**- Justification:** PR #10 modified the Go version and was closed. The project relies on specific toolchain versions pinned in `mise.toml` to ensure consistent builds across environments. Arbitrary version changes cause environment drift.
**- Files Affected:** `go.mod`, `mise.toml`

## IGNORE: Removing Dogfooding Wrappers

**- Pattern:** Removing the `go run ./cmd/ci-status run ... --` wrappers around commands in `mise.toml`.
**- Justification:** PR #30 was rejected. The `mise.toml` tasks (like `lint` and `test`) intentionally wrap execution with the tool itself to dogfood it during development. Removing these wrappers breaks the self-hosting setup.
**- Files Affected:** `mise.toml`

## IGNORE: Pinning golangci-lint to 2.x

**- Pattern:** Attempting to pin `golangci-lint` to version `2.x` (e.g., `2.9.0`) in `mise.toml`.
**- Justification:** PR #33 was rejected. `golangci-lint` version 2 does not exist; the '2' refers to the `.golangci.yml` configuration format version, not the tool version.
**- Files Affected:** `mise.toml`
