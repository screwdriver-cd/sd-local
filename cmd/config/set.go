package config

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

func newConfigSetCmd() *cobra.Command {
	configSetCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Short usage",
		Long:  `Long usage`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			isLocalOpt, err := cmd.Flags().GetBool("local")
			if err != nil {
				logrus.Fatal(err)
			}

			key := args[0]
			value := args[1]

			fmt.Println("Execute Set")
			fmt.Println("Local opt:", isLocalOpt)
			fmt.Printf("Key: %s\tValue: %s\n", key, value)
		},
	}

	return configSetCmd
}
