# Janitor's Journal

## 2025-02-14 - Initialize Journal

**Issue:** Missing journal file.
**Root Cause:** First run or file was deleted.
**Solution:** Created the journal file.
**Pattern:** Always check for required documentation files.

## 2025-02-14 - Refactor execute function in run command

**Issue:** The `execute` function in `cmd/ci-status/run.go` was calling `os.Exit` directly, making it hard to test and breaking the flow of the `RunE` cobra method.
**Root Cause:** The function was designed to be the final step of execution.
**Solution:** Changed `execute` to return `(int, error)` so the caller can decide how to exit.
**Pattern:** Avoid `os.Exit` in helper functions; return status and error instead.
