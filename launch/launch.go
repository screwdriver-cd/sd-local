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

// Launcher able to run local build
type Launcher interface {
	Run() error
}

var _ (Launcher) = (*launch)(nil)

type launch struct {
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
	ArtifactsPath string                 `json:"-"`
	MemoryLimit   string                 `json:"-"`
	SrcPath       string                 `json:"-"`
}

// Option is option for launch New
type Option struct {
	Job           screwdriver.Job
	Config        config.Config
	JobName       string
	JWT           string
	ArtifactsPath string
	Memory        string
	SrcPath       string
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

func createBuildConfig(option Option) buildConfig {
	defaultEnv := envVar{
		"SD_TOKEN":         option.JWT,
		"SD_ARTIFACTS_DIR": defaultArtDir,
		"SD_API_URL":       option.Config.APIURL,
		"SD_STORE_URL":     option.Config.StoreURL,
	}
	env := mergeEnv(defaultEnv, option.Job.Environment)

	return buildConfig{
		ID:            0,
		Environment:   env,
		EventID:       0,
		JobID:         0,
		ParentBuildID: []int{0},
		Sha:           "dummy",
		Meta:          map[string]interface{}{},
		Steps:         option.Job.Steps,
		Image:         option.Job.Image,
		JobName:       option.JobName,
		ArtifactsPath: option.ArtifactsPath,
		MemoryLimit:   option.Memory,
		SrcPath:       option.SrcPath,
	}
}

// New creates new Launcher interface.
func New(option Option) Launcher {
	l := new(launch)

	l.runner = newDocker(option.Config.Launcher.Image, option.Config.Launcher.Version)
	l.buildConfig = createBuildConfig(option)

	return l
}

// Run runs the build specified.
func (l *launch) Run() error {
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
