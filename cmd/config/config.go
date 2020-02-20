package config

import (
	"log"

	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Short usage",
		Long:  `Long usage`,
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	configCmd.PersistentFlags().Bool("local", false, "Run command with .sdlocal/config file in current directory.")

	configCmd.AddCommand(
		newConfigSetCmd(),
		newConfigViewCmd(),
	)

	return configCmd
}
