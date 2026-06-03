// Package docker wraps the docker / docker compose CLIs via os/exec, staying
// true to the Docker workflow developers already use.
package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ComposeUp starts the compose project in detached mode, optionally building.
func ComposeUp(dir, composeFile string, build bool, envFile string) error {
	args := []string{"compose", "-f", composeFile}
	if envFile != "" {
		args = append(args, "--env-file", envFile)
	}
	args = append(args, "up", "-d")
	if build {
		args = append(args, "--build")
	}
	return run(dir, "docker", args...)
}

// ComposeLogs streams logs for the compose project (-f follow optional).
func ComposeLogs(dir, composeFile string, follow bool) error {
	args := []string{"compose", "-f", composeFile, "logs"}
	if follow {
		args = append(args, "-f")
	}
	cmd := exec.Command("docker", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// ComposeDown stops and removes the compose project's resources.
func ComposeDown(dir, composeFile string) error {
	return run(dir, "docker", "compose", "-f", composeFile, "down")
}

// Build builds an image from the Dockerfile in dir and tags it.
func Build(dir, dockerfile, tag string) error {
	return run(dir, "docker", "build", "-f", dockerfile, "-t", tag, ".")
}

// Run starts a detached container mapping hostPort -> containerPort and returns
// the container ID.
func Run(dir, tag string, hostPort, containerPort int, envFile string) (string, error) {
	args := []string{"run", "-d", "-p", fmt.Sprintf("%d:%d", hostPort, containerPort)}
	if envFile != "" {
		args = append(args, "--env-file", envFile)
	}
	args = append(args, tag)
	cmd := exec.Command("docker", args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("docker run failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// RunLogs shows logs for a container started with docker run.
func RunLogs(containerID string, follow bool) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, containerID)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RemoveContainer force-removes a container by ID (ignores already-gone).
func RemoveContainer(id string) error {
	if id == "" {
		return nil
	}
	return run("", "docker", "rm", "-f", id)
}

func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
