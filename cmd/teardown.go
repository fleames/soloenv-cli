package cmd

import (
	"os"

	"github.com/fleames/soloenv-cli/internal/docker"
	"github.com/fleames/soloenv-cli/internal/state"
)

// teardown stops tunnel, auth proxy, watcher, app, and clears state.
func teardown(dir string, st *state.State, tun interface{ Stop() }) {
	if st == nil {
		_ = state.Remove(dir)
		return
	}

	if tun != nil {
		tun.Stop()
	} else if st.TunnelPID > 0 {
		if p, err := os.FindProcess(st.TunnelPID); err == nil {
			_ = p.Kill()
		}
	}

	if st.AuthProxyPID > 0 {
		if p, err := os.FindProcess(st.AuthProxyPID); err == nil {
			_ = p.Kill()
		}
	}

	if st.WatcherPID > 0 {
		if p, err := os.FindProcess(st.WatcherPID); err == nil {
			_ = p.Kill()
		}
	}

	workDir := dir
	if st.ProjectDir != "" {
		workDir = st.ProjectDir
	}

	switch st.Mode {
	case state.ModeCompose:
		if err := docker.ComposeDown(workDir, st.ComposeFile); err != nil {
			warn("failed to stop compose project: %v", err)
		}
	case state.ModeDockerfile:
		if err := docker.RemoveContainer(st.ContainerID); err != nil {
			warn("failed to remove container: %v", err)
		}
	}

	if err := state.Remove(workDir); err != nil {
		warn("failed to clear state: %v", err)
	}
}
