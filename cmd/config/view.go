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
		Short: "Short usage",
		Long:  `Long usage`,
		Run: func(cmd *cobra.Command, args []string) {
			isLocalOpt, err := cmd.Flags().GetBool("local")
			if err != nil {
				logrus.Fatal(err)
			}
			path, err := filePath(isLocalOpt)
			if err != nil {
				logrus.Fatal(err)
			}
			c, err := configNew(path)
			if err != nil {
				logrus.Fatal(err)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 5, 2, 2, ' ', 0)
			if isLocalOpt {
				fmt.Fprintln(w, "KEY\tVALUE")
				fmt.Fprintln(w, "api-url\tlocal.api.screwdriver.com")
				fmt.Fprintln(w, "store-url\tlocal.store.screwdriver.com")
				fmt.Fprintln(w, "token\tlocal.token")
				fmt.Fprintln(w, "launcher-version\tlatest")
				fmt.Fprintln(w, "launcher-image\tscrewdrivercd/launcher")
			} else {
				fmt.Fprintln(w, "KEY\tVALUE")
				fmt.Fprintf(w, "api-url\t%s\n", c.APIURL)
				fmt.Fprintf(w, "store-url\t%s\n", c.StoreURL)
				fmt.Fprintf(w, "token\t%s\n", c.Token)
				fmt.Fprintf(w, "launcher-version\t%s\n", c.Launcher.Version)
				fmt.Fprintf(w, "launcher-image\t%s\n", c.Launcher.Image)
			}

			w.Flush()
		},
	}

	return configViewCmd
}
