package config

import (
	"strings"

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
* Screwdriver.cd UUID as "UUID"
* Screwdriver.cd launcher image as "launcher-image"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			key, value := args[0], args[1]

			path, err := filePath()
			if err != nil {
				return err
			}

			config, err := configNew(path)
			if err != nil {
				return err
			}

			entry, err := config.Entry(config.Current)
			if err != nil {
				return err
			}

			err = entry.Set(key, value)
			if err != nil {
				if isInvalidKeyError(err) {
					err := cmd.Help()
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}

			err = config.Save()
			if err != nil {
				return err
			}
			return nil
		},
	}

	return configSetCmd
}
