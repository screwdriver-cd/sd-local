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

func filePath(isLocalOpt bool) (string, error) {
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
