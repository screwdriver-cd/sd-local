package config

import (
	"fmt"
	"text/tabwriter"

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

			entry, err := config.Entry(config.Current)
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 5, 2, 2, ' ', 0)

			fmt.Fprintln(w, "KEY\tVALUE")
			fmt.Fprintf(w, "api-url\t%s\n", entry.APIURL)
			fmt.Fprintf(w, "store-url\t%s\n", entry.StoreURL)
			fmt.Fprintf(w, "token\t%s\n", entry.Token)
			fmt.Fprintf(w, "launcher-version\t%s\n", entry.Launcher.Version)
			fmt.Fprintf(w, "launcher-image\t%s\n", entry.Launcher.Image)

			w.Flush()
			return nil
		},
	}

	return configViewCmd
}
