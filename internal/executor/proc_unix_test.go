//go:build unix

package executor_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"ci-status/internal/executor"
)

// TestTimeoutKillsProcessGroup ensures a timed-out shell does not leave its
// grandchild (sleep) running. CommandContext alone only kills the shell PID.
func TestTimeoutKillsProcessGroup(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")

	// Shell starts sleep in the background, writes its PID, then waits.
	// After timeout the executor must kill the whole group so sleep is gone.
	script := `sleep 60 & echo $! >"$1"; wait`
	e := executor.New()
	e.Stdout = &bytes.Buffer{}
	e.Stderr = &bytes.Buffer{}

	exitCode, err := e.Run(context.Background(), 200*time.Millisecond, "sh", []string{"-c", script, "sh", pidFile})
	if !errors.Is(err, executor.ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got exit=%d err=%v", exitCode, err)
	}
	if exitCode != executor.ExitCodeTimeout {
		t.Fatalf("expected exit code %d, got %d", executor.ExitCodeTimeout, exitCode)
	}

	// Allow a brief moment for the pid file to appear (written before wait).
	deadline := time.Now().Add(2 * time.Second)
	var childPID int
	for {
		data, readErr := os.ReadFile(pidFile)
		if readErr == nil {
			pid, convErr := strconv.Atoi(strings.TrimSpace(string(data)))
			if convErr == nil && pid > 0 {
				childPID = pid
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("child pid file never written: %v", readErr)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Process should be dead: Signal(0) fails with ESRCH.
	deadline = time.Now().Add(2 * time.Second)
	for {
		err := syscall.Kill(childPID, 0)
		if err != nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("grandchild sleep (pid %d) still alive after process-group kill", childPID)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
