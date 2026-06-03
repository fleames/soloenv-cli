// Package state persists the details of a running SoloEnv environment so that
// later commands (down, status) can find and clean it up.
package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const dirName = ".soloenv"
const fileName = "state.json"

const (
	ModeCompose    = "compose"
	ModeDockerfile = "dockerfile"
)

// State is the on-disk record of a live environment.
type State struct {
	ProjectDir   string     `json:"project_dir"`
	URL          string     `json:"url"`
	TunnelPID    int        `json:"tunnel_pid"`
	TunnelPort   int        `json:"tunnel_port"` // port cloudflared targets (may be auth proxy)
	AppPort      int        `json:"app_port"`
	AuthProxyPID int        `json:"auth_proxy_pid,omitempty"`
	AuthEnabled  bool       `json:"auth_enabled"`
	AuthUser     string     `json:"auth_user,omitempty"`
	Detached     bool       `json:"detached"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	WatcherPID   int        `json:"watcher_pid,omitempty"`
	Mode         string     `json:"mode"`
	ComposeFile  string     `json:"compose_file,omitempty"`
	ContainerID  string     `json:"container_id,omitempty"`
	ImageTag     string     `json:"image_tag,omitempty"`
	StartedAt    time.Time  `json:"started_at"`
}

func Dir(projectDir string) string  { return filepath.Join(projectDir, dirName) }
func Path(projectDir string) string { return filepath.Join(Dir(projectDir), fileName) }

func Exists(projectDir string) bool {
	_, err := os.Stat(Path(projectDir))
	return err == nil
}

func Save(projectDir string, s *State) error {
	if err := os.MkdirAll(Dir(projectDir), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(projectDir), data, 0o644)
}

func Load(projectDir string) (*State, error) {
	data, err := os.ReadFile(Path(projectDir))
	if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func Remove(projectDir string) error {
	return os.RemoveAll(Dir(projectDir))
}

// Expired reports whether the environment passed its TTL deadline.
func (s *State) Expired() bool {
	if s == nil || s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}
