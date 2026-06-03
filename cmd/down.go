package cmd

import (
	"github.com/spf13/cobra"
)

var flagDir string

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Tear down the running environment",
	RunE:  runDown,
}

func runDown(_ *cobra.Command, args []string) error {
	dir, err := projectDir(flagDir)
	if err != nil {
		return err
	}
	st, err := loadState(dir)
	if err != nil {
		return err
	}
	if st == nil {
		info("No SoloEnv environment is running here.")
		return nil
	}
	info("Tearing down...")
	teardown(dir, st, nil)
	success("Done. Environment is gone.")
	return nil
}
