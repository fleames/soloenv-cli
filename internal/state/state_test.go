package state

import (
	"testing"
	"time"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	exp := time.Now().Add(time.Hour).UTC()
	in := &State{
		ProjectDir:  dir,
		URL:         "https://calm-river-1234.trycloudflare.com",
		TunnelPID:   4242,
		TunnelPort:  54321,
		AppPort:     8080,
		AuthEnabled: true,
		AuthUser:    "solo",
		Detached:    true,
		ExpiresAt:   &exp,
		Mode:        ModeCompose,
		ComposeFile: "compose.yaml",
		StartedAt:   time.Now().UTC(),
	}

	if err := Save(dir, in); err != nil {
		t.Fatal(err)
	}
	if !Exists(dir) {
		t.Fatal("Exists = false after Save")
	}

	out, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if out.URL != in.URL || out.TunnelPID != in.TunnelPID || out.AppPort != in.AppPort ||
		out.Mode != in.Mode || out.ComposeFile != in.ComposeFile || !out.AuthEnabled || !out.Detached {
		t.Fatalf("roundtrip mismatch: %+v", out)
	}
	if out.ExpiresAt == nil || !out.ExpiresAt.Equal(exp) {
		t.Fatalf("ExpiresAt mismatch: %v vs %v", out.ExpiresAt, exp)
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, &State{URL: "x"}); err != nil {
		t.Fatal(err)
	}
	if err := Remove(dir); err != nil {
		t.Fatal(err)
	}
	if Exists(dir) {
		t.Fatal("Exists = true after Remove")
	}
}

func TestExpired(t *testing.T) {
	past := time.Now().Add(-time.Minute)
	future := time.Now().Add(time.Minute)

	if (&State{}).Expired() {
		t.Fatal("no deadline should not be expired")
	}
	if !(&State{ExpiresAt: &past}).Expired() {
		t.Fatal("past deadline should be expired")
	}
	if (&State{ExpiresAt: &future}).Expired() {
		t.Fatal("future deadline should not be expired")
	}
	var nilState *State
	if nilState.Expired() {
		t.Fatal("nil state should not be expired")
	}
}

func TestLoadMissing(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatal("expected error loading missing state")
	}
}
