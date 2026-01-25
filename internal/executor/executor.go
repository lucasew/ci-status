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
// It returns the command's exit code and an error if one occurred.
//
// Exit Codes:
// - 0: Command succeeded.
// - 124: Command timed out.
// - Other: Command failed with that exit code.
//
// If the command fails to start, it returns 0 and an error.
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
