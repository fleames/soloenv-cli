package cmd

import "testing"

func TestSanitize(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"my-app", "my-app"},
		{"My App", "my-app"},
		{"weird!!name", "weird--name"},
		{"UPPER_case", "upper_case"},
		{"...", "app"},
		{"", "app"},
		{"café", "caf"},
		{"123", "123"},
	}
	for _, tt := range tests {
		if got := sanitize(tt.in); got != tt.want {
			t.Fatalf("sanitize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
