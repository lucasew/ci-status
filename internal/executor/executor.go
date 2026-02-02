package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// Executor is responsible for running system commands with configured I/O streams.
// By default, it writes to os.Stdout and os.Stderr.
type Executor struct {
	Stdout io.Writer
	Stderr io.Writer
}

// New creates a default Executor writing to standard output and error.
func New() *Executor {
	return &Executor{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Run executes a command with the specified arguments and an optional timeout.
// It manages the command lifecycle, handling startup, execution, and timeout signals.
//
// Parameters:
// - ctx: The parent context. If cancelled, the command will be killed.
// - timeout: If > 0, creates a derived context with this timeout.
// - command: The executable name or path.
// - args: Arguments for the command.
//
// Returns:
// - int: The exit code of the command (0 for success, 124 for timeout, or actual exit code).
// - error: An error object if the command failed to start or timed out.
//
// Note: If the command fails to start (e.g. executable not found), it returns exit code 0 and an error.
// This distinguishes execution failures from application failures.
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
			// 124 is a standard exit code for timeout in GNU coreutils
			return 124, fmt.Errorf("command timed out")
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), nil
		}
		return 1, fmt.Errorf("command execution failed: %w", err)
	}

	return 0, nil
}
