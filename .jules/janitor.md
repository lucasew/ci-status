## 2026-02-02 - Refactor set command to use centralized forge initialization

### Issue
The `cmd/ci-status/set.go` file contains duplicated logic for detecting the forge client and commit SHA. This logic is already implemented in the helper function `initForge` in `cmd/ci-status/common.go`.

### Root Cause
Redundant implementation of initialization logic, likely due to parallel development or evolution of the helper function without updating all call sites.

### Solution
Refactor `set.go` to use `initForge`, removing the manual `DetectClient` and `DetectCommit` calls and their error handling. This reduces code duplication and ensures consistent initialization across commands.

### Pattern
DRY (Don't Repeat Yourself). Centralize common initialization logic.
