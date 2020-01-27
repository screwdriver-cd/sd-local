package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"honnef.co/go/tools/config"
)

func newBuildCmd(c config.Config, api screwdriver.API) *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build [job name]",
		Short: "Run screwdriver build of the specified job name.",
		Long:  `Run screwdriver build of the specified job name.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			job, err := api.Job("./screwdriver.yaml")
			if err != nil {
				return err
			}

			err := launch.Launch(job, c)
			if err != nil {
				return err
			}

			return fmt.Errorf("something happen")
		},
	}
	return buildCmd
}
