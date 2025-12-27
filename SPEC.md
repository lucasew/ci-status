
# ci-status CLI Specification

## Overview
`ci-status` is a cross-platform CLI tool that wraps command execution and automatically reports status to different forge platforms (GitHub, GitLab, Bitbucket). It detects the forge context automatically and reports pending/success/failure states based on command exit codes.

## Project Setup
- **Language:** Go
- **Dependency Management:** Use [Mise-en-place](https://mise.jdx.dev/) for all tool and dependency management
- **Installation:** Download mise installation script to a temporary folder
- **Build System:** All builds must use `mise` commands

## Core Interface

### Forge Adapter Pattern
```go
type ForgeClient interface {
    SetStatus(ctx context.Context, opts StatusOpts) error
}

type StatusOpts struct {
    Commit      string
    Context     string
    State       State
    Description string
    TargetURL   string // optional
}

type State string

const (
    StateRunning State = "running"
    StatePending State = "pending"
    StateSuccess State = "success"
    StateFailure State = "failure"
    StateError   State = "error"
)
```

### Implementation Priority
1. **GitHub** (initial focus) - Use GitHub Status API (commit statuses)
2. GitLab (future)
3. Bitbucket (future)

## CLI Command Structure

### Main Command
```bash
ci-status run [flags] <context-name> -- <command> [args...]
```

### Arguments
- `<context-name>`: Required. The name that appears in the forge UI (e.g., "Lint", "Tests", "Build")
- `<command>`: Required after `--`. The command to execute
- `[args...]`: Optional. Arguments passed to the command

### Flags
```
--forge string
    Override automatic forge detection
    Values: github|gitlab|bitbucket
    Default: auto-detect from git remote

--commit string
    Override commit SHA
    Default: auto-detect from CI environment variables or git

--pr string
    Override pull request number
    Default: auto-detect from CI environment variables

--url string
    Target URL for "Details" link in forge UI
    Default: empty

--pending-desc string
    Description shown while command is running
    Default: "Running..."

--success-desc string
    Description shown when command exits with code 0
    Default: "Passed"

--failure-desc string
    Description shown when command exits with non-zero code
    Default: "Failed"

--timeout duration
    Maximum time allowed for command execution
    Format: Go duration (e.g., 5m, 1h30m, 30s)
    Default: no timeout

--silent
    Suppress output when running in noop mode or on errors
    Default: false

--sentry-monitor string
    Sentry Cron Monitor slug. If set, reports "in_progress" at start and "ok"/"error" at end.
    Requires SENTRY_DSN environment variable.
```

## Execution Flow

1. **Parse arguments and flags**
2. **Detect context:**
   - Forge type (from git remote or `--forge`)
   - Commit SHA (from CI env vars, git, or `--commit`)
   - PR number if applicable (from CI env vars or `--pr`)
3. **Check credentials:**
   - GitHub: `GITHUB_TOKEN` env var
   - If no credentials found → **noop mode** (execute command without reporting)
4. **Set pending status:**
   - State: `pending`
   - Description: value from `--pending-desc`
5. **Execute command:**
   - Stream stdout/stderr to terminal in real-time
   - Apply `--timeout` if specified
6. **Set final status:**
   - Exit code 0 → `success` with `--success-desc`
   - Exit code != 0 → `failure` with `--failure-desc`
   - Timeout → `error` with description "Timed out"
7. **Exit with same code as wrapped command**

## Auto-detection Logic

### Forge Detection
- Parse `git remote -v` output
- Match patterns: `github.com`, `gitlab.com`, `bitbucket.org`
- If no match and no `--forge` flag → noop mode

### Commit SHA Detection (priority order)
1. `--commit` flag if provided
2. CI environment variables:
   - GitHub Actions: `GITHUB_SHA`
   - GitLab CI: `CI_COMMIT_SHA`
   - Bitbucket Pipelines: `BITBUCKET_COMMIT`
3. `git rev-parse HEAD`
4. If none found → noop mode

### PR Detection (priority order)
1. `--pr` flag if provided
2. CI environment variables:
   - GitHub Actions: `GITHUB_REF` (extract PR number)
   - GitLab CI: `CI_MERGE_REQUEST_IID`
   - Bitbucket Pipelines: `BITBUCKET_PR_ID`
3. If none found → treat as commit status (not PR-specific)

### GitHub Token Detection
- Environment variable: `GITHUB_TOKEN`
- If not found → noop mode

## Noop Mode Behavior
When context cannot be detected or credentials are missing:
- Execute the command normally
- Do not attempt to set any status
- Optionally log warning to stderr (unless `--silent`)
- Exit with command's exit code

## Error Handling
- Invalid flags → print usage and exit 1
- Missing context name → print usage and exit 1
- Missing command after `--` → print usage and exit 1
- Command timeout → set error status, exit with code 124
- Forge API errors → log warning (unless `--silent`), still exit with command's exit code
- Sentry errors → log warning (unless silent), still exit with command's exit code.
  - Sentry initialization fails -> Warning, proceed without Sentry.
  - SENTRY_DSN missing when --sentry-monitor is used -> Exit 1.

## Examples

```bash
# Basic usage
ci-status run "Lint" -- eslint .

# With timeout
ci-status run "Tests" --timeout 5m -- pytest

# With custom descriptions
ci-status run "Build" \
  --pending-desc "Building application..." \
  --success-desc "Build completed ✓" \
  --failure-desc "Build failed ✗" \
  -- make build

# With target URL
ci-status run "Deploy" --url "$CI_JOB_URL" -- ./deploy.sh

# Override detection
ci-status run "Security Scan" \
  --forge github \
  --commit abc123 \
  -- trivy scan .
```

## GitHub Implementation Details

### API Endpoint
Use GitHub Status API: `POST /repos/{owner}/{repo}/statuses/{sha}`

### Authentication
- Use token from `GITHUB_TOKEN` environment variable
- Header: `Authorization: Bearer ${GITHUB_TOKEN}`

### Repository Detection
- Parse from git remote URL
- Extract owner and repo name
- Handle both SSH and HTTPS formats

### Status Mapping
- `pending` → `"state": "pending"`
- `success` → `"state": "success"`
- `failure` → `"state": "failure"`
- `error` → `"state": "error"`

### Request Body
```json
{
  "state": "success",
  "target_url": "https://...",
  "description": "Passed",
  "context": "ci/lint"
}
```

## Project Structure Suggestion
```
ci-status/
├── .mise.toml              # Mise configuration
├── main.go                 # CLI entrypoint
├── cmd/
│   └── run.go             # run subcommand
├── internal/
│   ├── forge/
│   │   ├── forge.go       # Interface definition
│   │   ├── github.go      # GitHub implementation
│   │   └── detect.go      # Auto-detection logic
│   ├── executor/
│   │   └── executor.go    # Command execution with timeout
│   └── config/
│       └── config.go      # Flag parsing and config
└── README.md
```

## Success Criteria
- Single binary, easy to distribute
- Works seamlessly in CI and locally (noop when appropriate)
- Clear error messages
- Exit codes match wrapped command
- Real-time command output streaming
