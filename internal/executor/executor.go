package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type Executor struct {
	Stdout io.Writer
	Stderr io.Writer
}

func New() *Executor {
	return &Executor{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

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
