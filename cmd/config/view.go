package config

import (
	"fmt"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
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
		Run: func(cmd *cobra.Command, args []string) {
			path, err := filePath()
			if err != nil {
				logrus.Fatal(err)
			}

			config, err := configNew(path)
			if err != nil {
				logrus.Fatal(err)
			}

			c, err := config.Get(config.Current)
			if err != nil {
				logrus.Fatal(err)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 5, 2, 2, ' ', 0)

			fmt.Fprintln(w, "KEY\tVALUE")
			fmt.Fprintf(w, "api-url\t%s\n", c.APIURL)
			fmt.Fprintf(w, "store-url\t%s\n", c.StoreURL)
			fmt.Fprintf(w, "token\t%s\n", c.Token)
			fmt.Fprintf(w, "launcher-version\t%s\n", c.Launcher.Version)
			fmt.Fprintf(w, "launcher-image\t%s\n", c.Launcher.Image)

			w.Flush()
		},
	}

	return configViewCmd
}
