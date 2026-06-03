package cmd

import (
	"os"
	"path/filepath"

	"github.com/fleames/soloenv-cli/internal/state"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&flagDir, "dir", "", "project directory (default: current)")
}

// projectDir returns the directory containing .soloenv state.
// Uses --dir flag when set, otherwise the current working directory.
func projectDir(flagDir string) (string, error) {
	if flagDir != "" {
		abs, err := filepath.Abs(flagDir)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	return os.Getwd()
}

func loadState(dir string) (*state.State, error) {
	if !state.Exists(dir) {
		return nil, nil
	}
	return state.Load(dir)
}
