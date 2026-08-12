//go:build unix

package executor

import (
	"os/exec"
	"syscall"
)

// prepareCommand puts the child in a new process group so killCommand can
// signal the whole tree (shell + helpers) on timeout/cancel.
func prepareCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killCommand terminates the process group. Negative PID targets the group
// created by Setpgid; Process.Kill is a fallback if the group is already gone.
// Both signals are best-effort: the child may have exited between Wait racing
// with cancel, in which case ESRCH (or equivalent) is expected and ignored.
func killCommand(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err == nil {
		return
	}
	// Group kill failed (already reaped, or Setpgid never stuck); try the child.
	if err := cmd.Process.Kill(); err != nil {
		// Process is gone — nothing left to signal.
		return
	}
}
