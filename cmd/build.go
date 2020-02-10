package cmd

import (
	"context"
	"log"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/screwdriver-cd/sd-local/buildlog"
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build [job name]",
		Short: "Run screwdriver build of the specified job name.",
		Long:  `Run screwdriver build of the specified job name.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			homedir, err := homedir.Dir()
			if err != nil {
				log.Fatal(err)
			}

			config, err := config.ReadConfig(path.Join(homedir, ".sdlocal", "config"))
			if err != nil {
				log.Fatal(err)
			}

			api, err := screwdriver.New(config.APIURL, config.Token)
			if err != nil {
				log.Fatal(err)
			}

			job, err := api.Job(args[0], "./screwdriver.yaml")
			if err != nil {
				log.Fatal(err)
			}

			log, err := buildlog.New(context.Background(), path.Join(launch.ArtifactsDir, launch.LogFile), os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
			go log.Run()

			// lauch.Newには API interfaceを渡すように修正したい。
			launch := launch.New(job, config, args[0], api.JWT())

			err = launch.Run()
			if err != nil {
			}

			log.Stop()

			return
		},
	}
	return buildCmd
}
