package config

import (
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	configNew = config.New
)

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
			isLocalOpt, err := cmd.Flags().GetBool("local")
			if err != nil {
				logrus.Fatal(err)
			}

			key := args[0]
			value := args[1]

			path, err := filePath(isLocalOpt)
			if err != nil {
				logrus.Fatal(err)
			}

			conf, err := configNew(path)
			if err != nil {
				logrus.Fatal(err)
			}

			err = conf.Set(key, value)
			if err != nil {
				logrus.Fatal(err)
			}
		},
	}

	return configSetCmd
}
