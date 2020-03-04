package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/screwdriver-cd/sd-local/buildlog"
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/screwdriver-cd/sd-local/scm"
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
	memory       = ""
	scmNew       = scm.New
	osMkdirAll   = os.MkdirAll
)

func mergeEnvFromFile(optionEnv *map[string]string, envFilePath string) error {
	absEnvFilePath, err := filepath.Abs(envFilePath)
	if err != nil {
		return err
	}

	env, err := godotenv.Read(absEnvFilePath)
	if err != nil {
		return fmt.Errorf("failed to read env file in `%s`: %v", absEnvFilePath, err)
	}

	for k, v := range env {
		if _, ok := (*optionEnv)[k]; !ok {
			(*optionEnv)[k] = v
		}
	}
	return nil
}

func newBuildCmd() *cobra.Command {
	var srcURL string
	var optionEnv map[string]string
	var envFilePath string

	buildCmd := &cobra.Command{
		Use:   "build [job name]",
		Short: "Run screwdriver build.",
		Long:  `Run screwdriver build of the specified job name.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if envFilePath != "" {
				err := mergeEnvFromFile(&optionEnv, envFilePath)
				if err != nil {
					logrus.Fatal(err)
				}
			}

			homedir, err := homedir.Dir()
			if err != nil {
				logrus.Fatal(err)
			}

			sdlocalDir := filepath.Join(homedir, ".sdlocal")
			cwd, err := os.Getwd()
			if err != nil {
				logrus.Fatal(err)
			}
			srcPath := cwd

			if srcURL != "" {
				logrus.Infof("Pulling the source code from %s...", srcURL)

				scm, err := scmNew(sdlocalDir, srcURL)
				if err != nil {
					logrus.Fatal(err)
				}
				defer func() {
					err = scm.Clean()
					if err != nil {
						logrus.Fatal(err)
					}
				}()

				err = scm.Pull()
				if err != nil {
					logrus.Fatal(err)
				}
				srcPath = scm.LocalPath()
			}

			config, err := configNew(filepath.Join(sdlocalDir, "config"))
			if err != nil {
				logrus.Fatal(err)
			}

			api, err := apiNew(config.APIURL, config.Token)
			if err != nil {
				logrus.Fatal(err)
			}

			jobName := args[0]

			sdYAMLPath := filepath.Join(srcPath, "screwdriver.yaml")
			job, err := api.Job(jobName, sdYAMLPath)
			if err != nil {
				logrus.Fatal(err)
			}

			artifactsPath, err := filepath.Abs(artifactsDir)
			if err != nil {
				logrus.Fatal(err)
			}

			err = osMkdirAll(artifactsPath, 0777)
			if err != nil {
				logrus.Fatal(err)
			}
			logger, err := buildLogNew(filepath.Join(artifactsPath, launch.LogFile), os.Stdout)
			if err != nil {
				logrus.Fatal(err)
			}
			go logger.Run()

			option := launch.Option{
				Job:           job,
				Config:        config,
				JobName:       jobName,
				JWT:           api.JWT(),
				ArtifactsPath: artifactsPath,
				Memory:        memory,
				SrcPath:       srcPath,
				OptionEnv:     optionEnv,
			}

			launch := launchNew(option)

			logrus.Info("Prepare to start build...")
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

	buildCmd.Flags().StringVarP(
		&memory,
		"memory",
		"m",
		"",
		"Memory limit for build container, which take a positive integer, followed by a suffix of b, k, m, g.")

	buildCmd.Flags().StringVar(
		&srcURL,
		"src-url",
		"",
		`Specify the source url to build.
ex) git@github.com:<org>/<repo>.git[#<branch>]
    https://github.com/<org>/<repo>.git[#<branch>]`)

	buildCmd.Flags().StringToStringVarP(
		&optionEnv,
		"env",
		"e",
		map[string]string{},
		"Set key and value relationship which is set as environment variables of Build Container. (<key>=<value>)",
	)

	buildCmd.Flags().StringVar(
		&envFilePath,
		"env-file",
		"",
		"Path to config file of environment variables. '.env' format file can be used.")

	return buildCmd
}
