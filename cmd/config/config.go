package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const (
	configFileName = "config"
	configDirName  = ".sdlocal"
)

var filePath = func(isLocalOpt bool) (string, error) {
	if isLocalOpt {
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(pwd, configDirName, configFileName), nil
	}
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDirName, configFileName), nil
}

// NewConfigCmd return config command.
func NewConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage settings related to sd-local.",
		Long:  `Manage settings related to sd-local.`,
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
