package cmd

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/screwdriver-cd/sd-local/buildlog"
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	waitIO = 1
)

func newBuildCmd() *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build [job name]",
		Short: "Run screwdriver build.",
		Long:  `Run screwdriver build of the specified job name.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			homedir, err := homedir.Dir()
			if err != nil {
				logrus.Fatal(err)
			}

			config, err := config.ReadConfig(path.Join(homedir, ".sdlocal", "config"))
			if err != nil {
				logrus.Fatal(err)
			}

			api, err := screwdriver.New(config.APIURL, config.Token)
			if err != nil {
				logrus.Fatal(err)
			}

			jobName := args[0]

			cwd, err := os.Getwd()
			if err != nil {
				logrus.Fatal(err)
			}
			sdYAMLPath := path.Join(cwd, "screwdriver.yaml")
			job, err := api.Job(jobName, sdYAMLPath)
			if err != nil {
				logrus.Fatal(err)
			}

			artifactsPath := path.Join(cwd, launch.ArtifactsDir)
			err = os.MkdirAll(artifactsPath, 0666)
			if err != nil {
				logrus.Fatal(err)
			}
			logger, err := buildlog.New(context.Background(), path.Join(artifactsPath, launch.LogFile), os.Stdout)
			if err != nil {
				logrus.Fatal(err)
			}
			go logger.Run()

			launch := launch.New(job, config, jobName, api.JWT())

			err = launch.Run()
			if err != nil {
				logrus.Fatal(err)
			}

			// Wait for I/O processing.
			time.Sleep(time.Second * waitIO)
			logger.Stop()

			return
		},
	}
	return buildCmd
}
