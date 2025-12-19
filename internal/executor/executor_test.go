package executor_test

import (
	"bytes"
	"context"
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
	if err.Error() != "command timed out" {
		t.Errorf("expected 'command timed out', got '%v'", err)
	}
	if exitCode != 124 {
		t.Errorf("expected exit code 124, got %d", exitCode)
	}
}
