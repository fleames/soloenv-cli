//go:build windows

package cmd

import "os"

// signalZero reports liveness on Windows. os.FindProcess only succeeds when the
// process handle could be opened, so reaching here means it is alive.
func signalZero(p *os.Process) bool {
	return p != nil
}
