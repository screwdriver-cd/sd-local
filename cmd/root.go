package cmd

import (
	"github.com/spf13/cobra"
	"honnef.co/go/tools/config"
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
func Execute(c config.Config, api screwdriver.api) error {
	rootCmd := newRootCmd()
	rootCmd.AddCommand(newBuildCmd(c, api))
	return rootCmd.Execute()
}
