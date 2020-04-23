package config

import (
	"github.com/spf13/cobra"
)

func newConfigUseCmd() *cobra.Command {
	configUseCmd := &cobra.Command{
		Use:   "use [name]",
		Short: "Use the config of sd-local",
		Long: `Use the specified config as current config.
You can confirm the current config in view sub command.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			name := args[0]

			path, err := filePath()
			if err != nil {
				return err
			}

			config, err := configNew(path)
			if err != nil {
				return err
			}

			err = config.SetCurrent(name)
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

	return configUseCmd
}
