package cmd

import (
	"fmt"
	"runtime"

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
			ver := fmt.Sprintf("%s\nplatform: %s/%s\ngo: %s\ncompiler: %s", version, runtime.GOOS, runtime.GOARCH, runtime.Version(), runtime.Compiler)
			fmt.Fprintln(cmd.OutOrStdout(), ver)
		},
	}
}
