package launch

import (
	"fmt"
	"os/exec"

	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/screwdriver"
)

var lookPath = exec.LookPath

type Runner interface {
	RunBuild(buildConfig BuildConfig) ([]byte, error)
	SetupBin() error
}

type Launch struct {
	buildConfig BuildConfig
	runner      Runner
}

type EnvVar map[string]string

type BuildConfig struct {
	ID            int                    `json:"id"`
	Environment   []EnvVar               `json:"environment"`
	EventID       int                    `json:"eventId"`
	JobID         int                    `json:"jobId"`
	ParentBuildID []int                  `json:"parentBuildId"`
	Sha           string                 `json:"sha"`
	Meta          map[string]interface{} `json:"meta"`
	Steps         []screwdriver.Step     `json:"steps"`
	Image         string                 `json:"-"`
	JobName       string                 `json:"-"`
}

type BuildEnvironment struct {
	SD_ARTIFACTS_DIR string
	SD_API_URL       string
	SD_STORE_URL     string
}

const (
	defaultArtDir = "/sd/workspace/artifacts"
)

func mergeEnv(env, userEnv EnvVar) []EnvVar {

	for k, v := range userEnv {
		env[k] = v
	}

	return []EnvVar{env}
}

func createBuildConfig(config config.Config, job screwdriver.Job, jobName, jwt string) BuildConfig {
	defaultEnv := EnvVar{
		"SD_TOKEN":         jwt,
		"SD_ARTIFACTS_DIR": defaultArtDir,
		"SD_API_URL":       config.APIURL,
		"SD_STORE_URL":     config.StoreURL,
	}
	env := mergeEnv(defaultEnv, job.Environment)

	return BuildConfig{
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

func New(job screwdriver.Job, config config.Config, jobName, jwt string) *Launch {
	l := new(Launch)

	l.runner = newDocker(config.Launcher.Image, config.Launcher.Version)
	l.buildConfig = createBuildConfig(config, job, jobName, jwt)

	return l
}

func (l *Launch) runBuild(image, jobName, apiURL, storeURL string) error {

	return nil
}

func (l *Launch) Run() error {
	if _, err := lookPath("docker"); err != nil {
		return fmt.Errorf("`docker` command is not found in $PATH: %v", err)
	}

	if err := l.runner.SetupBin(); err != nil {
		return fmt.Errorf("failed to setup build: %v", err)
	}

	out, err := l.runner.RunBuild(l.buildConfig)
	if err != nil {
		return fmt.Errorf("failed to run build: %v", err)
	}

	fmt.Println(string(out))
	return nil
}
