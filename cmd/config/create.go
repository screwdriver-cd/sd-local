package config

import (
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configNew = config.New
)

func newConfigCreateCmd() *cobra.Command {
	configCreateCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create the config of sd-local",
		Long: `Create the config of sd-local.
The new config has only launcher-version and launcher-image.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			path, err := filePath()
			if err != nil {
				logrus.Fatal(err)
			}

			config, err := configNew(path)
			if err != nil {
				logrus.Fatal(err)
			}

			err = config.AddEntry(name)
			if err != nil {
				logrus.Fatal(err)
			}

			err = config.Save()
			if err != nil {
				logrus.Fatal(err)
			}
		},
	}

	return configCreateCmd
}
