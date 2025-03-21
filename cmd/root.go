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

// Cleaner will post-process sd-local.
type Cleaner interface {
	Kill(os.Signal)
	Clean()
}

var (
	flagVerbose bool
)

func newRootCmd() *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "sd-local",
		Short: "Run build in local",
		Long: `Run build instantly on your local machine with
a mostly the same environment as Screwdriver.cd's`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Help()
			if err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVarP(
		&flagVerbose,
		"verbose",
		"v",
		false,
		"verbose output.")

	return rootCmd
}

func kill(sig os.Signal) {
	for _, v := range cleaners {
		v.Kill(sig)
	}
}

func clean() {
	for _, v := range cleaners {
		v.Clean()
	}
}

// Execute executes the root command.
func Execute() error {
	cleaners = make([]Cleaner, 0, 2)
	defer clean()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		sig := <-quit
		kill(sig)
		clean()
		os.Exit(1)
	}()

	rootCmd := newRootCmd()
	rootCmd.SilenceErrors = true
	rootCmd.AddCommand(
		newBuildCmd(),
		config.NewConfigCmd(),
		newVersionCmd(),
		newUpdateCmd(),
	)
	return rootCmd.Execute()
}
