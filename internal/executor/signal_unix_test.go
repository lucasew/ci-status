//go:build unix

package executor_test

import (
	"bytes"
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

// TestSignalKillsProcessGroup ensures SIGTERM to the parent cancels Run and
// kills the child's process group. Setpgid means the child would not get the
// terminal signal itself; without NotifyContext it would be left orphaned.
func TestSignalKillsProcessGroup(t *testing.T) {
	assertSignalKillsProcessGroup(t, syscall.SIGTERM)
}

// TestSIGHUPKillsProcessGroup ensures terminal hangup cancels Run and kills
// the child's process group. Without SIGHUP in interruptSignals, the default
// disposition terminates the parent while Setpgid children keep running.
func TestSIGHUPKillsProcessGroup(t *testing.T) {
	assertSignalKillsProcessGroup(t, syscall.SIGHUP)
}

func assertSignalKillsProcessGroup(t *testing.T, sig syscall.Signal) {
	t.Helper()
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")

	script := `sleep 60 & echo $! >"$1"; wait`
	e := executor.New()
	e.Stdout = &bytes.Buffer{}
	e.Stderr = &bytes.Buffer{}

	done := make(chan struct {
		code int
		err  error
	}, 1)
	go func() {
		code, err := e.Run(t.Context(), 0, "sh", []string{"-c", script, "sh", pidFile})
		done <- struct {
			code int
			err  error
		}{code, err}
	}()

	// Wait until the grandchild pid is known, then signal this process.
	deadline := time.Now().Add(3 * time.Second)
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

	if err := syscall.Kill(os.Getpid(), sig); err != nil {
		t.Fatalf("send %v: %v", sig, err)
	}

	select {
	case res := <-done:
		if res.err == nil {
			t.Fatalf("expected error after %v", sig)
		}
		if errors.Is(res.err, executor.ErrTimeout) {
			t.Fatalf("signal should not report timeout: %v", res.err)
		}
		if res.code != 1 {
			t.Fatalf("expected exit code 1 on signal cancel, got %d", res.code)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("Run did not return after %v", sig)
	}

	deadline = time.Now().Add(2 * time.Second)
	for {
		err := syscall.Kill(childPID, 0)
		if err != nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("grandchild sleep (pid %d) still alive after %v cancel", childPID, sig)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
