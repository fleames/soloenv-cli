package project

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ResolveEnvFile picks the env file for docker compose/run.
// Priority: flag > config > .env.staging > .env.soloenv
func ResolveEnvFile(dir string, cfg *Config, flagPath string) (string, error) {
	if flagPath != "" {
		p := flagPath
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("env file not found: %s", p)
		}
		return p, nil
	}
	if cfg != nil && cfg.EnvFile != "" {
		p := filepath.Join(dir, cfg.EnvFile)
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("env file from soloenv.yml not found: %s", cfg.EnvFile)
		}
		return p, nil
	}
	for _, name := range []string{".env.staging", ".env.soloenv"} {
		p := filepath.Join(dir, name)
		if fileExists(p) {
			return p, nil
		}
	}
	return "", nil
}

// ResolveAuth returns user and password for basic auth. Empty password means no auth.
func ResolveAuth(cfg *Config, flagPassword string, flagProtect bool) (user, password string, err error) {
	password = flagPassword
	if password == "" && cfg != nil {
		password = cfg.Password
	}
	if flagProtect && password == "" {
		password, err = randomPassword(10)
		if err != nil {
			return "", "", err
		}
	}
	user = "solo"
	if cfg != nil && cfg.AuthUser != "" {
		user = cfg.AuthUser
	}
	return user, password, nil
}

// ResolveTTL parses duration from flag or config (e.g. "2h", "45m").
func ResolveTTL(cfg *Config, flagTTL string) (time.Duration, error) {
	raw := flagTTL
	if raw == "" && cfg != nil {
		raw = cfg.TTL
	}
	if raw == "" {
		return 0, nil
	}
	return time.ParseDuration(raw)
}

func randomPassword(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}
