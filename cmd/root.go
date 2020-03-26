package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/screwdriver-cd/sd-local/cmd/config"
	"github.com/spf13/cobra"
)

var (
	cleaners []Cleaner
)

type Cleaner interface {
	Kill(os.Signal)
	Clean()
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sd-local",
		Short: "Run build in local",
		Long: `Run build instantly on your local machine with
a mostly the same environment as Screwdriver.cd's`,
	}
	return rootCmd
}

func kill(sig os.Signal) {
	for _, v := range cleaners {
		v.Kill(sig)
	}
}

func cleanExit(code int) {
	for _, v := range cleaners {
		v.Clean()
	}
	os.Exit(code)
}

// Execute executes the root command.
func Execute() error {
	cleaners = make([]Cleaner, 0, 2)
	defer cleanExit(1)

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		for {
			select {
			case sig := <-quit:
				kill(sig)
				cleanExit(1)
				return
			}
		}
	}()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(
		newBuildCmd(),
		config.NewConfigCmd(),
		newVersionCmd(),
	)
	return rootCmd.Execute()
}
