package cmd

import (
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current environment and its URL",
	RunE:  runStatus,
}

func runStatus(_ *cobra.Command, _ []string) error {
	dir, err := projectDir(flagDir)
	if err != nil {
		return err
	}
	st, err := loadState(dir)
	if err != nil {
		return err
	}
	if st == nil {
		info("No SoloEnv environment is running here.")
		info("Run `soloenv up` to start one.")
		return nil
	}

	if st.Expired() {
		warn("TTL expired. Run `soloenv down` to clean up.")
	}

	success("SoloEnv environment is running")
	color.New(color.FgCyan, color.Bold).Printf("  URL:      %s\n", st.URL)
	info("  Project:  %s", st.ProjectDir)
	info("  Mode:     %s", st.Mode)
	info("  App port: %d", st.AppPort)
	if st.AuthEnabled {
		info("  Auth:     basic (%s)", st.AuthUser)
	}
	if st.Detached {
		info("  Detached: yes")
	}
	info("  Uptime:   %s", time.Since(st.StartedAt).Round(time.Second))
	if st.ExpiresAt != nil {
		left := time.Until(*st.ExpiresAt).Round(time.Second)
		if left > 0 {
			info("  Expires:  in %s", left)
		} else {
			warn("  Expires:  overdue")
		}
	}
	if processAlive(st.TunnelPID) {
		info("  Tunnel:   running (pid %d)", st.TunnelPID)
	} else {
		warn("Tunnel (pid %d) not running — URL may be dead.", st.TunnelPID)
	}
	return nil
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return signalZero(p)
}
