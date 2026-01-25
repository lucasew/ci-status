package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// Executor handles command execution with timeout support and output capture.
type Executor struct {
	Stdout io.Writer
	Stderr io.Writer
}

// New creates a new Executor configured to write to os.Stdout and os.Stderr.
func New() *Executor {
	return &Executor{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Run executes a command with an optional timeout.
//
// It returns the exit code of the command and an error.
// If the command times out, it returns exit code 124 and an error "command timed out".
// If the command fails to start, it returns 0 and an error.
// If the command runs but returns a non-zero exit code, it returns that code and nil error.
func (e *Executor) Run(ctx context.Context, timeout time.Duration, command string, args []string) (int, error) {
	var cmd *exec.Cmd

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd = exec.CommandContext(ctx, command, args...)
	cmd.Stdout = e.Stdout
	cmd.Stderr = e.Stderr

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start command: %w", err)
	}

	err := cmd.Wait()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return 124, fmt.Errorf("command timed out")
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), nil
		}
		return 1, fmt.Errorf("command execution failed: %w", err)
	}

	return 0, nil
}
