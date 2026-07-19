package executor_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"ci-status/internal/executor"
)

func TestExecutorRun(t *testing.T) {
	e := executor.New()
	var stdout, stderr bytes.Buffer
	e.Stdout = &stdout
	e.Stderr = &stderr

	ctx := context.Background()

	// Test successful execution
	exitCode, err := e.Run(ctx, 0, "echo", []string{"hello"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	if strings.TrimSpace(stdout.String()) != "hello" {
		t.Errorf("expected stdout 'hello', got '%s'", stdout.String())
	}

	// Test failure execution
	stdout.Reset()
	exitCode, err = e.Run(ctx, 0, "false", nil)
	if err != nil {
		t.Fatalf("expected no error (just exit code), got %v", err)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	// Test timeout
	stdout.Reset()
	exitCode, err = e.Run(ctx, 100*time.Millisecond, "sleep", []string{"1"})
	if err == nil {
		t.Fatal("expected error for timeout, got nil")
	}
	if !errors.Is(err, executor.ErrTimeout) {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
	if exitCode != executor.ExitCodeTimeout {
		t.Errorf("expected exit code %d, got %d", executor.ExitCodeTimeout, exitCode)
	}
}

// TestExecutorPassesStdin ensures wrapped commands can read from the executor's
// Stdin. Before this, Stdin was left nil and the child always saw /dev/null,
// so pipelines like `echo hi | ci-status run t -- cat` produced empty output.
func TestExecutorPassesStdin(t *testing.T) {
	e := executor.New()
	var stdout bytes.Buffer
	e.Stdin = strings.NewReader("hello-from-stdin\n")
	e.Stdout = &stdout
	e.Stderr = &bytes.Buffer{}

	exitCode, err := e.Run(context.Background(), 0, "cat", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if got := stdout.String(); got != "hello-from-stdin\n" {
		t.Fatalf("expected stdin to reach the command, got %q", got)
	}
}

func TestNewInheritsProcessStdin(t *testing.T) {
	e := executor.New()
	if e.Stdin == nil {
		t.Fatal("New() must set Stdin so children do not read /dev/null by default")
	}
}

func TestCancelKillsCommand(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	e := executor.New()
	e.Stdout = &bytes.Buffer{}
	e.Stderr = &bytes.Buffer{}

	done := make(chan struct {
		code int
		err  error
	}, 1)
	go func() {
		code, err := e.Run(ctx, 0, "sleep", []string{"30"})
		done <- struct {
			code int
			err  error
		}{code, err}
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case res := <-done:
		if res.err == nil {
			t.Fatal("expected error after cancel")
		}
		if errors.Is(res.err, executor.ErrTimeout) {
			t.Fatalf("cancel should not report timeout: %v", res.err)
		}
		if res.code != 1 {
			t.Fatalf("expected exit code 1 on cancel, got %d", res.code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}
