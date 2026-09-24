//go:build unix

package executor_test

import (
	"errors"
	"testing"
	"time"

	"ci-status/internal/executor"
)

// TestTimeoutKillsProcessGroup ensures a timed-out shell does not leave its
// grandchild (sleep) running. CommandContext alone only kills the shell PID.
func TestTimeoutKillsProcessGroup(t *testing.T) {
	e, pidFile := grandchildSleepHarness(t)

	exitCode, err := e.Run(t.Context(), 200*time.Millisecond, "sh", []string{"-c", grandchildSleepScript, "sh", pidFile})
	if !errors.Is(err, executor.ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got exit=%d err=%v", exitCode, err)
	}
	if exitCode != executor.ExitCodeTimeout {
		t.Fatalf("expected exit code %d, got %d", executor.ExitCodeTimeout, exitCode)
	}

	childPID := waitForGrandchildPID(t, pidFile, 2*time.Second)
	waitUntilGrandchildGone(t, childPID, "process-group kill")
}
