package config

import (
	"fmt"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"
)

func newConfigViewCmd() *cobra.Command {
	configViewCmd := &cobra.Command{
		Use:   "view",
		Short: "View the config of sd-local.",
		Long: `View the config of sd-local.
Can see the below settings:
* Screwdriver.cd API URL
* Screwdriver.cd Store URL
* Screwdriver.cd Token
* Screwdriver.cd launcher version
* Screwdriver.cd launcher image`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			path, err := filePath()
			if err != nil {
				return err
			}

			config, err := configNew(path)
			if err != nil {
				return err
			}

			for name, entry := range config.Entries {
				if name == config.Current {
					fmt.Fprintf(cmd.OutOrStdout(), "* %s:\n", name)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s:\n", name)
				}

				yaml, err := yaml.Marshal(entry)
				if err != nil {
					return err
				}

				for _, line := range strings.Split(string(yaml), "\n") {
					if line != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", line)
					}
				}

			}

			return nil
		},
	}

	return configViewCmd
}
