//go:build windows

package executor

import (
	"os/exec"
	"strconv"
)

// prepareCommand is a no-op on Windows. Job-object containment is not wired
// yet; killCommand uses taskkill /T to tear down the process tree instead.
func prepareCommand(cmd *exec.Cmd) {}

// killCommand terminates the wrapped process and its descendants.
// Process.Kill alone only stops the direct child (unlike Unix process groups),
// so shells that spawn helpers would leave orphans on timeout/cancel.
// taskkill /T /F mirrors that tree-wide teardown; Process.Kill is the fallback
// if taskkill is unavailable or the PID is already gone.
func killCommand(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pid := strconv.Itoa(cmd.Process.Pid)
	_ = exec.Command("taskkill", "/T", "/F", "/PID", pid).Run()
	_ = cmd.Process.Kill()
}
