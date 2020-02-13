package cmd

import (
	"context"
	"log"
	"os"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/screwdriver-cd/sd-local/buildlog"
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/spf13/cobra"
)

const (
	WAIT_IO = 1
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

			cwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			artifactsPath := path.Join(cwd, launch.ArtifactsDir)
			err = os.MkdirAll(artifactsPath, 0666)
			if err != nil {
				log.Fatal(err)
			}
			logger, err := buildlog.New(context.Background(), path.Join(artifactsPath, launch.LogFile), os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
			go logger.Run()

			// lauch.Newには API interfaceを渡すように修正したい。
			launch := launch.New(job, config, args[0], api.JWT())

			err = launch.Run()
			if err != nil {
				log.Fatal(err)
			}

			// Wait for I/O processing.
			time.Sleep(time.Second * WAIT_IO)
			logger.Stop()

			return
		},
	}
	return buildCmd
}
