package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/fleames/soloenv-cli/internal/tunnel"
)

// startTunnelDetached runs cloudflared in a new session with logs on disk so the
// tunnel keeps running after soloenv up --detach exits.
func startTunnelDetached(bin string, port int, logPath string, timeout time.Duration) (publicURL string, pid int, err error) {
	targetURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", 0, err
	}

	cmd := exec.Command(bin, "tunnel", "--no-autoupdate", "--url", targetURL)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	applyDetached(cmd)
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return "", 0, err
	}
	_ = logFile.Close()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(250 * time.Millisecond)
		data, readErr := os.ReadFile(logPath)
		if readErr != nil {
			continue
		}
		if u := tunnel.PickTunnelURLFromText(string(data)); u != "" {
			return u, cmd.Process.Pid, nil
		}
	}

	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	return "", 0, fmt.Errorf("timed out after %s waiting for a tunnel URL", timeout)
}
