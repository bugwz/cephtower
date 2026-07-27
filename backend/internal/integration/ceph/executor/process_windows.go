//go:build windows

package executor

import "os/exec"

func configureCommandProcess(_ *exec.Cmd) {}

func killCommandProcess(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}
