package cmd

import (
	"fmt"

	"github.com/fleames/soloenv-cli/internal/docker"
	"github.com/fleames/soloenv-cli/internal/state"
	"github.com/spf13/cobra"
)

var flagFollow bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream logs from the running app",
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().BoolVarP(&flagFollow, "follow", "f", true, "follow log output")
}

func runLogs(cmd *cobra.Command, args []string) error {
	dir, err := projectDir(flagDir)
	if err != nil {
		return err
	}
	st, err := loadState(dir)
	if err != nil {
		return err
	}
	if st == nil {
		info("No environment running.")
		return nil
	}
	workDir := dir
	if st.ProjectDir != "" {
		workDir = st.ProjectDir
	}
	switch st.Mode {
	case state.ModeCompose:
		return docker.ComposeLogs(workDir, st.ComposeFile, flagFollow)
	case state.ModeDockerfile:
		return docker.RunLogs(st.ContainerID, flagFollow)
	default:
		return fmt.Errorf("unknown mode %q", st.Mode)
	}
}
