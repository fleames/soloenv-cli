package project

import (
	"path/filepath"
	"testing"
	"time"
)

func TestResolveTTL(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		flag    string
		want    time.Duration
		wantErr bool
	}{
		{"none", &Config{}, "", 0, false},
		{"flag", &Config{}, "2h", 2 * time.Hour, false},
		{"config", &Config{TTL: "45m"}, "", 45 * time.Minute, false},
		{"flag overrides config", &Config{TTL: "45m"}, "1h", time.Hour, false},
		{"invalid", &Config{}, "banana", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveTTL(tt.cfg, tt.flag)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("ResolveTTL = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveAuth(t *testing.T) {
	t.Run("no auth by default", func(t *testing.T) {
		_, pass, err := ResolveAuth(&Config{}, "", false)
		if err != nil {
			t.Fatal(err)
		}
		if pass != "" {
			t.Fatalf("expected empty password, got %q", pass)
		}
	})

	t.Run("explicit password", func(t *testing.T) {
		user, pass, err := ResolveAuth(&Config{}, "hunter2", false)
		if err != nil {
			t.Fatal(err)
		}
		if pass != "hunter2" || user != "solo" {
			t.Fatalf("got user=%q pass=%q", user, pass)
		}
	})

	t.Run("protect generates password", func(t *testing.T) {
		_, pass, err := ResolveAuth(&Config{}, "", true)
		if err != nil {
			t.Fatal(err)
		}
		if len(pass) < 8 {
			t.Fatalf("generated password too short: %q", pass)
		}
	})

	t.Run("custom user from config", func(t *testing.T) {
		user, _, err := ResolveAuth(&Config{AuthUser: "reviewer", Password: "x"}, "", false)
		if err != nil {
			t.Fatal(err)
		}
		if user != "reviewer" {
			t.Fatalf("user = %q, want reviewer", user)
		}
	})
}

func TestResolveEnvFile(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		dir := t.TempDir()
		got, err := ResolveEnvFile(dir, &Config{}, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Fatalf("expected empty, got %q", got)
		}
	})

	t.Run("auto-detect .env.staging", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, ".env.staging", "FOO=bar\n")
		got, err := ResolveEnvFile(dir, &Config{}, "")
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(got) != ".env.staging" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("flag wins and must exist", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "custom.env", "A=1\n")
		got, err := ResolveEnvFile(dir, &Config{}, "custom.env")
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(got) != "custom.env" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("flag missing errors", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := ResolveEnvFile(dir, &Config{}, "nope.env"); err == nil {
			t.Fatal("expected error for missing env file")
		}
	})

	t.Run("config env_file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, ".env.prod", "A=1\n")
		got, err := ResolveEnvFile(dir, &Config{EnvFile: ".env.prod"}, "")
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(got) != ".env.prod" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestRandomPasswordUnique(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		p, err := randomPassword(10)
		if err != nil {
			t.Fatal(err)
		}
		if len(p) != 10 {
			t.Fatalf("len = %d, want 10", len(p))
		}
		if seen[p] {
			t.Fatalf("duplicate password generated: %q", p)
		}
		seen[p] = true
	}
}
