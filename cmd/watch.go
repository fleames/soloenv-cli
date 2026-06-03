package cmd

import (
	"time"

	"github.com/fleames/soloenv-cli/internal/state"
	"github.com/spf13/cobra"
)

var watchDir string
var watchUntil string

var watchCmd = &cobra.Command{
	Use:    "watch",
	Hidden: true,
	Short:  "Internal: wait until deadline then tear down",
	RunE:   runWatch,
}

func init() {
	watchCmd.Flags().StringVar(&watchDir, "dir", "", "project directory")
	watchCmd.Flags().StringVar(&watchUntil, "until", "", "RFC3339 deadline")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(_ *cobra.Command, args []string) error {
	if watchDir == "" || watchUntil == "" {
		return nil
	}
	deadline, err := time.Parse(time.RFC3339, watchUntil)
	if err != nil {
		return err
	}
	wait := time.Until(deadline)
	if wait > 0 {
		time.Sleep(wait)
	}
	st, err := state.Load(watchDir)
	if err != nil || st == nil {
		return nil
	}
	if st.Expired() || time.Now().After(deadline) {
		info("TTL reached — tearing down...")
		teardown(watchDir, st, nil)
		success("Environment expired and was removed.")
	}
	return nil
}

func spawnWatcher(projectDir string, expiresAt time.Time) (int, error) {
	exe, err := executablePath()
	if err != nil {
		return 0, err
	}
	c := execWatchCommand(exe, projectDir, expiresAt)
	if err := c.Start(); err != nil {
		return 0, err
	}
	return c.Process.Pid, nil
}
