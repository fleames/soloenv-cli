//go:build !windows

package cmd

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

func execCommandDetached(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	applyDetached(cmd)
	return cmd
}

func applyDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func execWatchCommand(exe, projectDir string, expiresAt time.Time) *exec.Cmd {
	cmd := exec.Command(exe, "watch", "--dir", projectDir, "--until", expiresAt.Format(time.RFC3339))
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd
}
