package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/fleames/soloenv-cli/internal/authproxy"
	"github.com/spf13/cobra"
)

var (
	holdAppPort  int
	holdUser     string
	holdPassword string
)

var holdCmd = &cobra.Command{
	Use:    "hold",
	Hidden: true,
	Short:  "Internal: keep auth proxy running until killed",
	RunE:   runHold,
}

func init() {
	holdCmd.Flags().IntVar(&holdAppPort, "app-port", 0, "upstream app port")
	holdCmd.Flags().StringVar(&holdUser, "user", "solo", "basic auth user")
	holdCmd.Flags().StringVar(&holdPassword, "password", "", "basic auth password")
	rootCmd.AddCommand(holdCmd)
}

func runHold(cmd *cobra.Command, args []string) error {
	if holdAppPort == 0 || holdPassword == "" {
		return fmt.Errorf("hold requires --app-port and --password")
	}
	proxy, err := authproxy.Start(holdAppPort, holdUser, holdPassword)
	if err != nil {
		return err
	}
	// Port is written for parent via stdout (one line) for rare race-free handoff.
	fmt.Printf("%d\n", proxy.Port)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	proxy.Stop()
	return nil
}

func spawnAuthHold(appPort int, user, password string) (pid, proxyPort int, err error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, 0, err
	}
	c := exec.Command(exe, "hold",
		"--app-port", fmt.Sprintf("%d", appPort),
		"--user", user,
		"--password", password,
	)
	c.Stderr = nil
	c.Stdin = nil
	setDetached(c)
	stdout, err := c.StdoutPipe()
	if err != nil {
		return 0, 0, err
	}
	if err := c.Start(); err != nil {
		return 0, 0, err
	}
	r := bufio.NewReader(stdout)
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return 0, 0, err
	}
	if _, err := fmt.Sscanf(line, "%d", &proxyPort); err != nil {
		return 0, 0, fmt.Errorf("auth proxy did not report port: %w", err)
	}
	time.Sleep(50 * time.Millisecond)
	return c.Process.Pid, proxyPort, nil
}

func setDetached(c *exec.Cmd) {
	applyDetached(c)
}
