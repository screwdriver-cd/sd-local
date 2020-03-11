package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// version is embedded when building this command using ldflags.
	// if nothing is embedded, version is "dev"
	version = "dev"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display command's version.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}
