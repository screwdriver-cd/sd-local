package config

import (
	"github.com/screwdriver-cd/sd-local/config"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			name := args[0]

			path, err := filePath()
			if err != nil {
				return err
			}

			c, err := configNew(path)
			if err != nil {
				return err
			}

			defaultEntry := config.DefaultEntry()
			err = c.AddEntry(name, defaultEntry)
			if err != nil {
				return err
			}

			err = c.Save()
			if err != nil {
				return err
			}
			return nil
		},
	}

	return configCreateCmd
}
