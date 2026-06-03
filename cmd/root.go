// Package cmd defines the SoloEnv CLI commands.
package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var version = "0.2.2"

var rootCmd = &cobra.Command{
	Use:     "soloenv",
	Short:   "One-command ephemeral staging for solo devs",
	Long:    "SoloEnv runs your existing Docker app and gives you a shareable public staging URL,\nthen tears it down when you are done. No cloud account, no DevOps detour.",
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(upCmd, downCmd, statusCmd, openCmd, logsCmd)
}

func filepathBase(p string) string { return filepath.Base(p) }

func sanitize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-_")
	if out == "" {
		return "app"
	}
	return out
}
