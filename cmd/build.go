package cmd

import (
	"os"
	"path/filepath"
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

var (
	configNew    = config.New
	apiNew       = screwdriver.New
	buildLogNew  = buildlog.New
	launchNew    = launch.New
	artifactsDir = launch.ArtifactsDir
)

func newBuildCmd() *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build [job name]",
		Short: "Run screwdriver build.",
		Long:  `Run screwdriver build of the specified job name.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			homedir, err := homedir.Dir()
			if err != nil {
				logrus.Fatal(err)
			}

			config, err := configNew(filepath.Join(homedir, ".sdlocal", "config"))
			if err != nil {
				logrus.Fatal(err)
			}

			api, err := apiNew(config.APIURL, config.Token)
			if err != nil {
				logrus.Fatal(err)
			}

			jobName := args[0]

			cwd, err := os.Getwd()
			if err != nil {
				logrus.Fatal(err)
			}

			sdYAMLPath := filepath.Join(cwd, "screwdriver.yaml")
			job, err := api.Job(jobName, sdYAMLPath)
			if err != nil {
				logrus.Fatal(err)
			}

			artifactsPath, err := filepath.Abs(artifactsDir)
			if err != nil {
				logrus.Fatal(err)
			}

			err = os.MkdirAll(artifactsPath, 0777)
			if err != nil {
				logrus.Fatal(err)
			}
			logger, err := buildLogNew(filepath.Join(artifactsPath, launch.LogFile), os.Stdout)
			if err != nil {
				logrus.Fatal(err)
			}
			go logger.Run()

			launch := launchNew(job, config, jobName, api.JWT(), artifactsPath)

			err = launch.Run()
			if err != nil {
				logrus.Fatal(err)
			}

			// Wait for I/O processing.
			time.Sleep(time.Second * waitIO)
			logger.Stop()
		},
	}

	buildCmd.Flags().StringVar(
		&artifactsDir,
		"artifacts-dir",
		launch.ArtifactsDir,
		"Path to the host side directory which is mounted into $SD_ARTIFACTS_DIR.")

	return buildCmd
}
