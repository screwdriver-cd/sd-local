package config

import (
	"fmt"

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

			fmt.Println("Execute View")
			fmt.Println("Local opt:", isLocalOpt)
		},
	}

	return configViewCmd
}
