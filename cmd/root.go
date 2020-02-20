package cmd

import (
	"github.com/screwdriver-cd/sd-local/cmd/config"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sd-local",
		Short: "Run build in local",
		Long: `Run build instantly on your local machine with
a mostly the same environment as Screwdriver.cd's`,
	}
	return rootCmd
}

// Execute executes the root command.
func Execute() error {
	rootCmd := newRootCmd()
	rootCmd.AddCommand(
		newBuildCmd(),
		config.NewConfigCmd(),
	)
	return rootCmd.Execute()
}
