package launch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/stretchr/testify/assert"
)

var testDir string = "./testdata"

func newBuildConfig() buildConfig {
	buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
	job := screwdriver.Job{}
	json.Unmarshal(buf, &job)
	return buildConfig{
		ID: 0,
		Environment: []envVar{{
			"SD_ARTIFACTS_DIR": "/test/artifacts",
			"SD_API_URL":       "http://api-test.screwdriver.cd",
			"SD_STORE_URL":     "http://store-test.screwdriver.cd",
			"SD_TOKEN":         "testjwt",
			"FOO":              "foo",
		}},
		EventID:       0,
		JobID:         0,
		ParentBuildID: []int{0},
		Sha:           "dummy",
		Meta:          map[string]interface{}{},
		Steps:         job.Steps,
		Image:         job.Image,
		JobName:       "test",
	}
}

func TestNew(t *testing.T) {
	t.Run("success with custom artifacts dir", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		json.Unmarshal(buf, &job)
		job.Environment["SD_ARTIFACTS_DIR"] = "/test/artifacts"

		config := config.Config{
			APIURL:   "http://api-test.screwdriver.cd",
			StoreURL: "http://store-test.screwdriver.cd",
			Token:    "testtoken",
			Launcher: config.Launcher{Version: "latest", Image: "screwdrivercd/launcher"},
		}

		expectedBuildConfig := newBuildConfig()

		l := New(job, config, "test", "testjwt")

		assert.Equal(t, expectedBuildConfig, l.buildConfig)
	})

	t.Run("success with default artifacts dir", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		json.Unmarshal(buf, &job)

		config := config.Config{
			APIURL:   "http://api-test.screwdriver.cd",
			StoreURL: "http://store-test.screwdriver.cd",
			Token:    "testtoken",
			Launcher: config.Launcher{Version: "latest", Image: "screwdrivercd/launcher"},
		}

		expectedBuildConfig := newBuildConfig()
		expectedBuildConfig.Environment[0]["SD_ARTIFACTS_DIR"] = "/sd/workspace/artifacts"

		l := New(job, config, "test", "testjwt")

		assert.Equal(t, expectedBuildConfig, l.buildConfig)
	})
}

type mockRunner struct {
	errorRunBuild error
	errorSetupBin error
}

func (m *mockRunner) RunBuild(buildConfig buildConfig) error {
	return m.errorRunBuild
}

func (m *mockRunner) SetupBin() error {
	return m.errorSetupBin
}

func TestRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		json.Unmarshal(buf, &job)

		launch := Launch{
			buildConfig: newBuildConfig(),
			runner: &mockRunner{
				errorRunBuild: nil,
				errorSetupBin: nil,
			},
		}

		lookPath = func(cmd string) (string, error) {
			return "/bin/docker", nil
		}

		defer func() {
			lookPath = exec.LookPath
		}()

		err := launch.Run()

		assert.Equal(t, nil, err)
	})

	t.Run("failure in lookPath", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		json.Unmarshal(buf, &job)

		launch := Launch{
			buildConfig: newBuildConfig(),
			runner: &mockRunner{
				errorRunBuild: nil,
				errorSetupBin: nil,
			},
		}

		lookPath = func(cmd string) (string, error) {
			return "", fmt.Errorf("exec: \"docker\": executable file not found in $PATH")
		}

		defer func() {
			lookPath = exec.LookPath
		}()

		err := launch.Run()

		assert.Equal(t, fmt.Errorf("`docker` command is not found in $PATH: exec: \"docker\": executable file not found in $PATH"), err)
	})

	t.Run("failure in SetupBin", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		json.Unmarshal(buf, &job)

		launch := Launch{
			buildConfig: newBuildConfig(),
			runner: &mockRunner{
				errorRunBuild: nil,
				errorSetupBin: fmt.Errorf("docker: Error response from daemon"),
			},
		}

		lookPath = func(cmd string) (string, error) {
			return "/bin/docker", nil
		}

		defer func() {
			lookPath = exec.LookPath
		}()

		err := launch.Run()

		assert.Equal(t, fmt.Errorf("failed to setup build: docker: Error response from daemon"), err)
	})

	t.Run("failure in RunBuild", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		json.Unmarshal(buf, &job)

		launch := Launch{
			buildConfig: newBuildConfig(),
			runner: &mockRunner{
				errorRunBuild: fmt.Errorf("docker: Error response from daemon"),
				errorSetupBin: nil,
			},
		}

		lookPath = func(cmd string) (string, error) {
			return "/bin/docker", nil
		}

		defer func() {
			lookPath = exec.LookPath
		}()

		err := launch.Run()

		assert.Equal(t, fmt.Errorf("failed to run build: docker: Error response from daemon"), err)
	})
}
