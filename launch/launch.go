package launch

import (
	"fmt"
	"os/exec"

	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/screwdriver"
)

var lookPath = exec.LookPath

type runner interface {
	runBuild(buildConfig buildConfig) error
	setupBin() error
}

// Launch has config to run build.
type Launch struct {
	buildConfig buildConfig
	runner      runner
}

type envVar map[string]string

type buildConfig struct {
	ID            int                    `json:"id"`
	Environment   []envVar               `json:"environment"`
	EventID       int                    `json:"eventId"`
	JobID         int                    `json:"jobId"`
	ParentBuildID []int                  `json:"parentBuildId"`
	Sha           string                 `json:"sha"`
	Meta          map[string]interface{} `json:"meta"`
	Steps         []screwdriver.Step     `json:"steps"`
	Image         string                 `json:"-"`
	JobName       string                 `json:"-"`
}

const (
	defaultArtDir = "/sd/workspace/artifacts"
)

func mergeEnv(env, userEnv envVar) []envVar {
	for k, v := range userEnv {
		env[k] = v
	}

	return []envVar{env}
}

func createBuildConfig(config config.Config, job screwdriver.Job, jobName, jwt string) buildConfig {
	defaultEnv := envVar{
		"SD_TOKEN":         jwt,
		"SD_ARTIFACTS_DIR": defaultArtDir,
		"SD_API_URL":       config.APIURL,
		"SD_STORE_URL":     config.StoreURL,
	}
	env := mergeEnv(defaultEnv, job.Environment)

	return buildConfig{
		ID:            0,
		Environment:   env,
		EventID:       0,
		JobID:         0,
		ParentBuildID: []int{0},
		Sha:           "dummy",
		Meta:          map[string]interface{}{},
		Steps:         job.Steps,
		Image:         job.Image,
		JobName:       jobName,
	}
}

// New creates new Launch struct.
func New(job screwdriver.Job, config config.Config, jobName, jwt string) *Launch {
	l := new(Launch)

	l.runner = newDocker(config.Launcher.Image, config.Launcher.Version)
	l.buildConfig = createBuildConfig(config, job, jobName, jwt)

	return l
}

// Run runs the build specified.
func (l *Launch) Run() error {
	if _, err := lookPath("docker"); err != nil {
		return fmt.Errorf("`docker` command is not found in $PATH: %v", err)
	}

	if err := l.runner.setupBin(); err != nil {
		return fmt.Errorf("failed to setup build: %v", err)
	}

	err := l.runner.runBuild(l.buildConfig)
	if err != nil {
		return fmt.Errorf("failed to run build: %v", err)
	}

	return nil
}
