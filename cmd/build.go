package cmd

import (
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/spf13/cobra"
)

func newBuildCmd(c config.Config, api screwdriver.API, l launch.Launcher) *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build [job name]",
		Short: "Run screwdriver build of the specified job name.",
		Long:  `Run screwdriver build of the specified job name.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return l.Run()
		},
	}
	return buildCmd
}
