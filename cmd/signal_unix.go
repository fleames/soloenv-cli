//go:build !windows

package cmd

import (
	"os"
	"syscall"
)

// signalZero probes a process with signal 0 to check liveness on Unix.
func signalZero(p *os.Process) bool {
	return p.Signal(syscall.Signal(0)) == nil
}
