package config

import (
	"github.com/spf13/cobra"
)

func newConfigDeleteCmd() *cobra.Command {
	configDeleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete the config of sd-local",
		Long:  `Delete the config of sd-local.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			path, err := filePath()
			if err != nil {
				return err
			}

			config, err := configNew(path)
			if err != nil {
				return err
			}

			err = config.DeleteEntry(name)
			if err != nil {
				return err
			}

			err = config.Save()
			if err != nil {
				return err
			}
			return nil
		},
	}

	return configDeleteCmd
}
