package cmd

import "github.com/spf13/cobra"

var (
	rootCmd = &cobra.Command{
		Use:   "sd-local",
		Short: "Able to build in local",
		Long: `Run build instantly on your local machine with
a mostly the same environment as Screwdriver.cd's`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
