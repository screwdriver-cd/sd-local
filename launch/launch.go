package launch

import (
	"encoding/json"
	"fmt"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"os"
	"os/exec"
	"strings"
)

type Launch struct {
	config LaunchConfig
}

type EnvVar map[string]string

type LaunchConfig struct {
	ID            int                    `json:"id"`
	Environment   []EnvVar               `json:"environment"`
	EventID       int                    `json:"eventId"`
	JobID         int                    `json:"jobId"`
	ParentBuildID []int                  `json:"parentBuildId"`
	Sha           string                 `json:"sha"`
	Meta          map[string]interface{} `json:"meta"`
	Steps         []screwdriver.Step                 `json:"steps"`
}

type BuildEnvironemnt struct {
	SD_ARTIFACTS_DIR string
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

func createLaunchConfig(job screwdriver.Job, token string) LaunchConfig {
	t1 := []EnvVar{EnvVar{"SD_TOKEN": token}}
	env := envMerge(t1, job.Environment)

	return LaunchConfig{
		ID: 0,
		Environment: env,
		EventID: 0,
		JobID: 0,
		ParentBuildID: []int{0},
		Sha: "dummy",
		Meta: map[string]interface{}{},
		Steps: job.Steps,
	}
}

func checkExecCmd(c string) (ok bool, err error) {
	cmd := exec.Command("which", c)
	err = cmd.Run()

	if err != nil {
		return false, err
	}

	status := cmd.ProcessState.ExitCode()

	if status == 0 {
		ok = true
	} else {
		ok = false
	}

	return ok, nil
}

func New(job screwdriver.Job, token string) *Launch {
	l := new(Launch)

	l.config = createLaunchConfig(job, token)

	return l
}

func runDocker(env BuildEnvironemnt, config LaunchConfig, image, jobName, apiURL, storeURL string) error {
	// 厳密にするならカレントかつscrewdriver.yamlがある場所にした方が良さそう
	cwd, err := os.Getwd()

	if err != nil {
		return nil
	}

	srcDir := cwd
	hostArtDir := cwd
	containerArtDir := env.SD_ARTIFACTS_DIR
	buildImage := image

	srcOpt := "-v" + srcDir + "/:/sd/workspace"
	artOpt := "-v" + hostArtDir + "/:/" + containerArtDir
	configJson, err := json.Marshal(config)

	if err != nil {
		return err
	}


	cmd := []string{"/opt/sd/local_run.sh", string(configJson), jobName, apiURL, storeURL, containerArtDir}
	execCmd := strings.Join(cmd, " ")


	out, err := exec.Command("docker", "run", "-d", "--rm", srcOpt, artOpt, buildImage, execCmd).Output()

	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}

func (l *Launch) Run() {
	_, err := checkExecCmd("ls")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	env := BuildEnvironemnt{SD_ARTIFACTS_DIR: "/sd/workspace/artifacts"}

	runDocker(env, l.config, "alpine", "main", "https://api-cd.screwdriver.corp.yahoo.co.jp", "https://store-cd.screwdriver.corp.yahoo.co.jp")
}