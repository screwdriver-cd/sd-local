package cmd

import (
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sd-local",
		Short: "Able to build in local",
		Long: `Run build instantly on your local machine with
a mostly the same environment as Screwdriver.cd's`,
	}
	return rootCmd
}

// Execute executes the root command.
func Execute(c config.Config, api screwdriver.API, l launch.Launcher) error {
	rootCmd := newRootCmd()
	rootCmd.AddCommand(newBuildCmd(c, api, l))
	return rootCmd.Execute()
}
