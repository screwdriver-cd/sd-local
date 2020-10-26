package cmd

import (
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	runCmd := newBuildCmdImpl(true)
	runCmd.Use = "run [job name]"
	runCmd.Short = "Attach screwdriver build container."
	runCmd.Long = `Attach screwdriver build container of the specified job name.`

	return runCmd
}
