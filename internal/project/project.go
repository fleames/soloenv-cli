// Package project detects the kind of Docker project in a directory, loads the
// optional soloenv.yml config, and resolves which local port to expose.
package project

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Kind is the type of Docker project detected.
type Kind int

const (
	// KindNone means no recognizable Docker project was found.
	KindNone Kind = iota
	// KindCompose means a Docker Compose file is present.
	KindCompose
	// KindDockerfile means a lone Dockerfile is present.
	KindDockerfile
)

// Project describes the detected Docker project.
type Project struct {
	Dir         string
	Kind        Kind
	ComposeFile string // filename relative to Dir (compose only)
	Dockerfile  string // filename relative to Dir (dockerfile only)
}

// Config is the optional soloenv.yml file in the project directory.
type Config struct {
	Port     int    `yaml:"port"`
	Service  string `yaml:"service"`
	Build    *bool  `yaml:"build"`
	EnvFile  string `yaml:"env_file"`
	Password string `yaml:"password"`
	AuthUser string `yaml:"auth_user"`
	TTL      string `yaml:"ttl"`
}

var composeNames = []string{"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}
var dockerfileNames = []string{"Dockerfile", "dockerfile"}

// Detect inspects dir and returns the detected Project. Compose takes priority
// over a lone Dockerfile, matching `docker compose` behavior.
func Detect(dir string) (*Project, error) {
	for _, name := range composeNames {
		if fileExists(filepath.Join(dir, name)) {
			return &Project{Dir: dir, Kind: KindCompose, ComposeFile: name}, nil
		}
	}
	for _, name := range dockerfileNames {
		if fileExists(filepath.Join(dir, name)) {
			return &Project{Dir: dir, Kind: KindDockerfile, Dockerfile: name}, nil
		}
	}
	return nil, fmt.Errorf("no compose file or Dockerfile found in %s\n  SoloEnv needs a compose.yaml/docker-compose.yml or a Dockerfile to run", dir)
}

// LoadConfig reads soloenv.yml/soloenv.yaml from dir if present. A missing file
// is not an error: it returns an empty config.
func LoadConfig(dir string) (*Config, error) {
	for _, name := range []string{"soloenv.yml", "soloenv.yaml"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var c Config
		if err := yaml.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", name, err)
		}
		return &c, nil
	}
	return &Config{}, nil
}

// composeConfig mirrors the relevant subset of `docker compose config --format json`.
type composeConfig struct {
	Name     string `json:"name"`
	Services map[string]struct {
		Ports []struct {
			Target    int `json:"target"`
			Published any `json:"published"`
		} `json:"ports"`
	} `json:"services"`
}

// ResolveComposePort determines the host port to expose for a compose project.
// override (from flag/config) wins if > 0. Otherwise it parses the rendered
// compose config: if service is given, its first published port is used; if not,
// and exactly one published port exists across services, that one is used.
func ResolveComposePort(dir, composeFile, service string, override int) (int, error) {
	if override > 0 {
		return override, nil
	}
	cmd := exec.Command("docker", "compose", "-f", composeFile, "config", "--format", "json")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("could not read compose config: %w", err)
	}
	var cfg composeConfig
	if err := json.Unmarshal(out, &cfg); err != nil {
		return 0, fmt.Errorf("could not parse compose config: %w", err)
	}

	type pub struct {
		service string
		port    int
	}
	var found []pub
	for name, svc := range cfg.Services {
		if service != "" && name != service {
			continue
		}
		for _, p := range svc.Ports {
			if port := toPort(p.Published); port > 0 {
				found = append(found, pub{service: name, port: port})
			}
		}
	}

	if service != "" {
		if len(found) == 0 {
			return 0, fmt.Errorf("service %q has no published ports; add a ports mapping or pass --port", service)
		}
		return found[0].port, nil
	}
	switch len(found) {
	case 0:
		return 0, fmt.Errorf("no published ports found in %s; add a ports mapping (e.g. \"8080:80\") or pass --port", composeFile)
	case 1:
		return found[0].port, nil
	default:
		var names []string
		for _, f := range found {
			names = append(names, fmt.Sprintf("%s:%d", f.service, f.port))
		}
		return 0, fmt.Errorf("multiple published ports found (%s); pick one with --port or --service", strings.Join(names, ", "))
	}
}

// DockerfileExpose reads the first EXPOSE port from the Dockerfile, or 0 if none.
func DockerfileExpose(dir, dockerfile string) int {
	f, err := os.Open(filepath.Join(dir, dockerfile))
	if err != nil {
		return 0
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(strings.ToUpper(line), "EXPOSE") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		// EXPOSE 8080/tcp -> 8080
		portPart := strings.SplitN(fields[1], "/", 2)[0]
		if port, err := strconv.Atoi(portPart); err == nil {
			return port
		}
	}
	return 0
}

// toPort coerces the JSON "published" value (string or number) into an int.
func toPort(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case string:
		// published may be "8080" or "127.0.0.1:8080".
		s := t
		if i := strings.LastIndex(s, ":"); i >= 0 {
			s = s[i+1:]
		}
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
	}
	return 0
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
