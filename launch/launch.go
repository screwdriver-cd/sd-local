package launch

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/sirupsen/logrus"
)

var (
	lookPath     = exec.LookPath
	apiVersion   = "v4"
	storeVersion = "v1"
)

type runner interface {
	runBuild(buildConfig buildConfig) error
	setupBin() error
	kill(os.Signal)
	clean()
}

// Launcher able to run local build
type Launcher interface {
	Run() error
	Kill(os.Signal)
	Clean()
}

var _ (Launcher) = (*launch)(nil)

type launch struct {
	buildConfig buildConfig
	runner      runner
}

// EnvVar is a map for environment variables
type EnvVar map[string]string

// Meta is a map for metadata
type Meta map[string]interface{}

type buildConfig struct {
	ID            int                `json:"id"`
	Environment   []EnvVar           `json:"environment"`
	EventID       int                `json:"eventId"`
	JobID         int                `json:"jobId"`
	ParentBuildID []int              `json:"parentBuildId"`
	Sha           string             `json:"sha"`
	Meta          Meta               `json:"meta"`
	Steps         []screwdriver.Step `json:"steps"`
	Image         string             `json:"-"`
	JobName       string             `json:"-"`
	ArtifactsPath string             `json:"-"`
	MemoryLimit   string             `json:"-"`
	SrcPath       string             `json:"-"`
	UseSudo       bool               `json:"-"`
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
	OptionEnv     EnvVar
	Meta          Meta
	UseSudo       bool
}

const (
	defaultArtDir = "/sd/workspace/artifacts"
)

func mergeEnv(env, jobEnv, optionEnv EnvVar) []EnvVar {
	for k, v := range jobEnv {
		env[k] = v
	}
	for k, v := range optionEnv {
		env[k] = v
	}

	return []EnvVar{env}
}

func createBuildConfig(option Option) buildConfig {
	apiURL, storeURL := option.Config.APIURL, option.Config.StoreURL

	a, err := url.Parse(option.Config.APIURL)
	if err == nil {
		a.Path = path.Join(a.Path, apiVersion)
		apiURL = a.String()
	} else {
		logrus.Warn("SD_API_URL is invalid. It may cause errors")
	}

	s, err := url.Parse(option.Config.StoreURL)
	if err == nil {
		s.Path = path.Join(s.Path, storeVersion)
		storeURL = s.String()
	} else {
		logrus.Warn("SD_STORE_URL is invalid. It may cause errors")
	}

	defaultEnv := EnvVar{
		"SD_TOKEN":         option.JWT,
		"SD_ARTIFACTS_DIR": defaultArtDir,
		"SD_API_URL":       apiURL,
		"SD_STORE_URL":     storeURL,
	}

	env := mergeEnv(defaultEnv, option.Job.Environment, option.OptionEnv)

	return buildConfig{
		ID:            0,
		Environment:   env,
		EventID:       0,
		JobID:         0,
		ParentBuildID: []int{0},
		Sha:           "dummy",
		Meta:          option.Meta,
		Steps:         option.Job.Steps,
		Image:         option.Job.Image,
		JobName:       option.JobName,
		ArtifactsPath: option.ArtifactsPath,
		MemoryLimit:   option.Memory,
		SrcPath:       option.SrcPath,
		UseSudo:       option.UseSudo,
	}
}

// New creates new Launcher interface.
func New(option Option) Launcher {
	l := new(launch)

	l.runner = newDocker(option.Config.Launcher.Image, option.Config.Launcher.Version, option.UseSudo)
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

func (l *launch) Kill(sig os.Signal) {
	l.runner.kill(sig)
}

func (l *launch) Clean() {
	l.runner.clean()
}
