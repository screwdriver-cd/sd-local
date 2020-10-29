package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

var (
	configNew     = config.New
	apiNew        = screwdriver.New
	buildLogNew   = buildlog.New
	launchNew     = launch.New
	artifactsDir  = launch.ArtifactsDir
	memory        = ""
	scmNew        = scm.New
	osMkdirAll    = os.MkdirAll
	useSudo       = false
	usePrivileged = false
	interactMode  = false
	loggerDone    chan struct{}
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
	var optionMeta string
	var metaFilePath string

	buildCmd := &cobra.Command{
		Use:   "build [job name]",
		Short: "Run screwdriver build.",
		Long:  `Run screwdriver build of the specified job name.`,
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)

			if err != nil {
				return err
			}

			if optionMeta != "" && metaFilePath != "" {
				return errors.New("can't pass the both options `meta` and `meta-file`, please specify only one of them")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cmd.SilenceUsage = true

			if envFilePath != "" {
				err = mergeEnvFromFile(&optionEnv, envFilePath)
				if err != nil {
					return err
				}
			}

			metaJSON := []byte("{}")
			if optionMeta != "" {
				metaJSON = []byte(optionMeta)
			} else if metaFilePath != "" {
				absMetaFilePath, err := filepath.Abs(metaFilePath)

				if err != nil {
					return err
				}

				metaJSON, err = ioutil.ReadFile(absMetaFilePath)

				if err != nil {
					return fmt.Errorf("failed to read meta-file %s: %v", metaFilePath, err)
				}
			}

			var meta launch.Meta

			err = json.Unmarshal(metaJSON, &meta)

			if err != nil {
				return fmt.Errorf("failed to parse meta %s, meta must be formated with JSON: %v", string(metaJSON), err)
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			configBaseDir, err := homedir.Dir()
			if err != nil {
				return err
			}

			sdlocalDir := filepath.Join(configBaseDir, ".sdlocal")
			srcPath := cwd

			if srcURL != "" {
				logrus.Infof("Pulling the source code from %s...", srcURL)

				scm, err := scmNew(sdlocalDir, srcURL, useSudo)
				if err != nil {
					return err
				}
				s, ok := scm.(Cleaner)
				if ok {
					cleaners = append(cleaners, s)
				}

				err = scm.Pull()
				if err != nil {
					return err
				}
				srcPath = scm.LocalPath()
			}

			config, err := configNew(filepath.Join(sdlocalDir, "config"))
			if err != nil {
				return err
			}

			entry, err := config.Entry(config.Current)
			if err != nil {
				return err
			}

			api, err := apiNew(entry.APIURL, entry.Token)
			if err != nil {
				return err
			}

			jobName := args[0]

			sdYAMLPath := filepath.Join(srcPath, "screwdriver.yaml")
			job, err := api.Job(jobName, sdYAMLPath)
			if err != nil {
				return err
			}

			artifactsPath, err := filepath.Abs(artifactsDir)
			if err != nil {
				return err
			}

			err = osMkdirAll(artifactsPath, 0777)
			if err != nil {
				return err
			}

			loggerDone = make(chan struct{})
			logger, err := buildLogNew(filepath.Join(artifactsPath, launch.LogFile), os.Stdout, loggerDone)
			if err != nil {
				return err
			}
			go logger.Run()

			option := launch.Option{
				Job:           job,
				Entry:         *entry,
				JobName:       jobName,
				JWT:           api.JWT(),
				ArtifactsPath: artifactsPath,
				Memory:        memory,
				SrcPath:       srcPath,
				OptionEnv:     optionEnv,
				Meta:          meta,
				UseSudo:       useSudo,
				UsePrivileged: usePrivileged,
				InteractMode:  interactMode,
				FlagVerbose:   flagVerbose,
			}

			launch := launchNew(option)
			l, ok := launch.(Cleaner)
			if ok {
				cleaners = append(cleaners, l)
			}

			logrus.Info("Prepare to start build...")
			err = launch.Run()
			if err != nil {
				return err
			}

			logger.Stop()
			<-loggerDone

			return nil
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

	buildCmd.Flags().StringVar(
		&optionMeta,
		"meta",
		"",
		"Metadata to pass into the build environment, which is represented with JSON format",
	)

	buildCmd.Flags().StringVar(
		&metaFilePath,
		"meta-file",
		"",
		"Path to the meta file. meta file is represented with JSON format.")

	buildCmd.Flags().BoolVar(
		&useSudo,
		"sudo",
		false,
		"Use sudo command for container runtime.")

	buildCmd.Flags().BoolVar(
		&usePrivileged,
		"privileged",
		false,
		"Use privileged mode for container runtime.")

	buildCmd.Flags().BoolVarP(
		&interactMode,
		"interact",
		"i",
		false,
		"Attach the build container in interactive mode.")

	return buildCmd
}
