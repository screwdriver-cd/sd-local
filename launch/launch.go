package launch

import (
	"fmt"
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"os"
	"os/exec"
)

type Runner interface {
	RunBuild(buildConfig BuildConfig, environment BuildEnvironment) ([]byte, error)
	SetupBin() error
}

type Launch struct {
	buildConfig BuildConfig
	buildEnvironment BuildEnvironment
	runner Runner
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
	Steps         []screwdriver.Step                 `json:"steps"`
	Image		  string				  `json:"-"`
	JobName		  string				  `json:"-"`
}

type BuildEnvironment struct {
	SD_ARTIFACTS_DIR string
	SD_API_URL string
	SD_STORE_URL string
}

func envMerge(env1 []EnvVar, env2 EnvVar) []EnvVar{
	merged := make([]EnvVar, 0)

	for _, v := range env1 {
		for k, vv := range v {
			merged = append(merged, EnvVar{k: vv})
		}
	}

	for k, v := range env2 {
		merged = append(merged, EnvVar{k: v})
	}

	return merged
}

func createBuildConfig(job screwdriver.Job, jobName, jwt string) BuildConfig {
	t1 := []EnvVar{EnvVar{"SD_TOKEN": jwt}}
	env := envMerge(t1, job.Environment)

	return BuildConfig{
		ID: 0,
		Environment: env,
		EventID: 0,
		JobID: 0,
		ParentBuildID: []int{0},
		Sha: "dummy",
		Meta: map[string]interface{}{},
		Steps: job.Steps,
		Image:job.Image,
		JobName: jobName,
	}
}

func New(job screwdriver.Job, config config.Config, jobName, jwt string) *Launch {
	l := new(Launch)

	l.runner = newDocker(config.Launcher.Image, config.Launcher.Version)
	l.buildConfig = createBuildConfig(job, jobName, jwt)
	l.buildEnvironment = BuildEnvironment{
		SD_ARTIFACTS_DIR: "/sd/workspace/artifacts",
		SD_API_URL: config.APIURL,
		SD_STORE_URL: config.StoreURL,
	}

	return l
}

func checkExecCmd(c string) (ok bool, err error) {
	cmd := exec.Command("which", c)
	err = cmd.Run()

	if err != nil {
		return false, fmt.Errorf("error: %s : also not exists command to %s.", err, c)
	}

	status := cmd.ProcessState.ExitCode()

	if status == 0 {
		ok = true
	} else {
		ok = false
	}

	return ok, nil
}

func (l *Launch) runBuild(image, jobName, apiURL, storeURL string) error {


	return nil
}

func (l *Launch) Run() {
	_, err := checkExecCmd("docker")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = l.runner.SetupBin()
	if err != nil {
		fmt.Println("SetupBin: ", err)
		os.Exit(1)
	}

	out, err := l.runner.RunBuild(l.buildConfig, l.buildEnvironment)
	if err != nil {
		fmt.Println("RunBuild: ", err)
		os.Exit(1)
	}

	fmt.Println(string(out))
}
