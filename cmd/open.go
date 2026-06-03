package cmd

import (
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the staging URL in your default browser",
	RunE:  runOpen,
}

func runOpen(cmd *cobra.Command, args []string) error {
	dir, err := projectDir(flagDir)
	if err != nil {
		return err
	}
	st, err := loadState(dir)
	if err != nil {
		return err
	}
	if st == nil || st.URL == "" {
		info("No environment running. Run `soloenv up` first.")
		return nil
	}
	return openURL(st.URL)
}

func openURL(url string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		c = exec.Command("open", url)
	default:
		c = exec.Command("xdg-open", url)
	}
	return c.Start()
}
