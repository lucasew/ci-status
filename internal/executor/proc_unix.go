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
func killCommand(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	_ = cmd.Process.Kill()
}
