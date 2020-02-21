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
The config of "sd-local" can be viewed in "sd-local config view".`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			return nil
		},
	}

	return configSetCmd
}
