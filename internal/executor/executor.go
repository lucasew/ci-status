package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// ErrTimeout is returned when a command exceeds its configured timeout.
// Callers should use errors.Is to detect timeouts rather than comparing error strings.
var ErrTimeout = errors.New("command timed out")

// ExitCodeTimeout is the process exit code used when a command times out.
// 124 matches the convention used by GNU timeout(1).
const ExitCodeTimeout = 124

// Executor is responsible for running system commands with configured I/O streams.
// By default, it inherits the process stdin/stdout/stderr so pipes and interactive
// input work when wrapping a command (e.g. `echo x | ci-status run t -- cat`).
type Executor struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// New creates a default Executor that inherits the process standard streams.
func New() *Executor {
	return &Executor{
		Stdin:  os.Stdin,
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
//
// On Unix, the child is started in its own process group so a timeout/cancel
// kills the whole tree (shells that spawn helpers no longer leave orphans).
// Because that new group no longer receives the terminal's SIGINT/SIGTERM with
// the parent, Run also cancels on those signals so killCommand still runs.
func (e *Executor) Run(ctx context.Context, timeout time.Duration, command string, args []string) (int, error) {
	// Setpgid isolates the child from the terminal process group. Without this,
	// Ctrl+C / SIGTERM would kill only ci-status and leave the wrapped command
	// (and its grandchildren) running.
	ctx, stop := withSignalCancel(ctx)
	defer stop()

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Own cancellation: CommandContext only kills the direct child PID, not a
	// process group. We set the group in prepareCommand and kill it ourselves.
	cmd := exec.Command(command, args...)
	// nil Stdin would make the child read from /dev/null, which breaks
	// pipelines and any command that expects inherited stdin.
	cmd.Stdin = e.Stdin
	cmd.Stdout = e.Stdout
	cmd.Stderr = e.Stderr
	prepareCommand(cmd)

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start command: %w", err)
	}

	waitErr := make(chan error, 1)
	go func() {
		waitErr <- cmd.Wait()
	}()

	select {
	case err := <-waitErr:
		if err == nil {
			return 0, nil
		}
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode(), nil
		}
		return 1, fmt.Errorf("command execution failed: %w", err)
	case <-ctx.Done():
		// Context cancelled or timed out. Kill the process group, then decide
		// from Wait's result (see outcomeAfterStop).
		killCommand(cmd)
		return outcomeAfterStop(<-waitErr, ctx.Err())
	}
}

// outcomeAfterStop maps a Wait result after timeout/cancel kill.
//
// select may choose the ctx.Done branch even when waitErr is also ready
// (both ready → random choice). If Wait already observed exit 0, prefer
// success over a false timeout/cancel. Non-nil Wait errors keep the
// stop reason (timeout vs cancel).
func outcomeAfterStop(waitErr, ctxErr error) (int, error) {
	if waitErr == nil {
		return 0, nil
	}
	if errors.Is(ctxErr, context.DeadlineExceeded) {
		return ExitCodeTimeout, ErrTimeout
	}
	return 1, fmt.Errorf("command cancelled: %w", ctxErr)
}
