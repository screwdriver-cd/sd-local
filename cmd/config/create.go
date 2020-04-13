package config

import (
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	newConfigList = config.NewConfigList
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

			configList, err := newConfigList(path)
			if err != nil {
				logrus.Fatal(err)
			}

			err = configList.Add(name)
			if err != nil {
				logrus.Fatal(err)
			}

			err = configList.Save()
			if err != nil {
				logrus.Fatal(err)
			}
		},
	}

	return configCreateCmd
}
