//go:build unix

package executor_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"ci-status/internal/executor"
)

// grandchildSleepScript starts sleep in the background, records its PID, then
// waits. The shell exiting is not enough: the grandchild must be reaped too.
const grandchildSleepScript = `sleep 60 & echo $! >"$1"; wait`

func grandchildSleepHarness(t *testing.T) (*executor.Executor, string) {
	t.Helper()
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	e := executor.New()
	e.Stdout = &bytes.Buffer{}
	e.Stderr = &bytes.Buffer{}
	return e, pidFile
}

// waitForGrandchildPID polls pidFile until it contains a positive PID.
func waitForGrandchildPID(t *testing.T, pidFile string, wait time.Duration) int {
	t.Helper()
	deadline := time.Now().Add(wait)
	var readErr error
	for {
		data, err := os.ReadFile(pidFile)
		readErr = err
		if err == nil {
			pid, convErr := strconv.Atoi(strings.TrimSpace(string(data)))
			if convErr == nil && pid > 0 {
				return pid
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("child pid file never written: %v", readErr)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// waitUntilGrandchildGone polls until Signal(0) fails. Both callers waited 2s.
// why keeps the fatal distinct ("process-group kill" vs "terminated cancel").
func waitUntilGrandchildGone(t *testing.T, pid int, why string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		err := syscall.Kill(pid, 0)
		if err != nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("grandchild sleep (pid %d) still alive after %s", pid, why)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
