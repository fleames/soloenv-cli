package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToPort(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int
	}{
		{"float", float64(8080), 8080},
		{"string plain", "8080", 8080},
		{"string with host", "127.0.0.1:8080", 8080},
		{"string ipv6ish", "0.0.0.0:3000", 3000},
		{"empty string", "", 0},
		{"garbage", "abc", 0},
		{"nil", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toPort(tt.in); got != tt.want {
				t.Fatalf("toPort(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestDockerfileExpose(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     int
	}{
		{"simple", "FROM nginx\nEXPOSE 8080\n", 8080},
		{"with proto", "FROM nginx\nEXPOSE 3000/tcp\n", 3000},
		{"lowercase keyword", "from nginx\nexpose 5000\n", 5000},
		{"indented", "FROM nginx\n   EXPOSE 9000\n", 9000},
		{"none", "FROM nginx\nCMD [\"nginx\"]\n", 0},
		{"first wins", "FROM nginx\nEXPOSE 80\nEXPOSE 443\n", 80},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(tt.contents), 0o644); err != nil {
				t.Fatal(err)
			}
			if got := DockerfileExpose(dir, "Dockerfile"); got != tt.want {
				t.Fatalf("DockerfileExpose = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	t.Run("compose preferred", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "compose.yaml", "services: {}\n")
		writeFile(t, dir, "Dockerfile", "FROM scratch\n")
		p, err := Detect(dir)
		if err != nil {
			t.Fatal(err)
		}
		if p.Kind != KindCompose {
			t.Fatalf("Kind = %v, want compose", p.Kind)
		}
		if p.ComposeFile != "compose.yaml" {
			t.Fatalf("ComposeFile = %q", p.ComposeFile)
		}
	})

	t.Run("dockerfile only", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "Dockerfile", "FROM scratch\n")
		p, err := Detect(dir)
		if err != nil {
			t.Fatal(err)
		}
		if p.Kind != KindDockerfile {
			t.Fatalf("Kind = %v, want dockerfile", p.Kind)
		}
	})

	t.Run("none", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := Detect(dir); err == nil {
			t.Fatal("expected error for empty dir")
		}
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("missing is empty", func(t *testing.T) {
		dir := t.TempDir()
		cfg, err := LoadConfig(dir)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Port != 0 || cfg.Service != "" {
			t.Fatalf("expected empty config, got %+v", cfg)
		}
	})

	t.Run("parses fields", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "soloenv.yml", "port: 8080\nservice: web\nenv_file: .env.staging\npassword: secret\nauth_user: reviewer\nttl: 4h\n")
		cfg, err := LoadConfig(dir)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Port != 8080 || cfg.Service != "web" || cfg.EnvFile != ".env.staging" ||
			cfg.Password != "secret" || cfg.AuthUser != "reviewer" || cfg.TTL != "4h" {
			t.Fatalf("unexpected config: %+v", cfg)
		}
	})
}

func writeFile(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
