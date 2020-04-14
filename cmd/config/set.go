package config

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func isInvalidKeyError(err error) bool {
	return strings.Contains(err.Error(), "invalid key")
}

func newConfigSetCmd() *cobra.Command {
	configSetCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set the config of sd-local",
		Long: `Set the config of sd-local.
Can set the below settings:
* Screwdriver.cd API URL as "api-url"
* Screwdriver.cd Store URL as "store-url"
* Screwdriver.cd Token as "token"
* Screwdriver.cd launcher version as "launcher-version"
* Screwdriver.cd launcher image as "launcher-image"`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key, value := args[0], args[1]

			path, err := filePath()
			if err != nil {
				logrus.Fatal(err)
			}

			configList, err := configNew(path)
			if err != nil {
				logrus.Fatal(err)
			}

			conf, err := configList.Get(configList.Current)
			if err != nil {
				logrus.Fatal(err)
			}

			err = conf.Set(key, value)
			if err != nil {
				if isInvalidKeyError(err) {
					err := cmd.Help()
					if err != nil {
						logrus.Fatal(err)
					}
				} else {
					logrus.Fatal(err)
				}
			}

			err = configList.Save()
			if err != nil {
				logrus.Fatal(err)
			}
		},
	}

	return configSetCmd
}
