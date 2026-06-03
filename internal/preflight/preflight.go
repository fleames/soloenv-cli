// Package preflight verifies the host has what SoloEnv needs before doing work.
package preflight

import (
	"fmt"
	"os/exec"
)

// CheckDocker ensures the docker CLI is installed and the daemon is reachable.
// It returns a user-facing error with a hint when something is missing.
func CheckDocker() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found on PATH.\n  Install Docker Desktop or Docker Engine: https://docs.docker.com/get-docker/")
	}
	// `docker info` fails fast (non-zero exit) when the daemon is not running.
	out, err := exec.Command("docker", "info", "--format", "{{.ServerVersion}}").CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker is installed but the daemon is not responding.\n  Start Docker and try again.\n  details: %s", trim(string(out)))
	}
	return nil
}

func trim(s string) string {
	if len(s) > 300 {
		return s[:300] + "..."
	}
	return s
}
