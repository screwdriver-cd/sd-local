package launch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/stretchr/testify/assert"
)

var testDir string = "./testdata"

func newBuildEntry(options ...func(b *buildEntry)) buildEntry {
	buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
	job := screwdriver.Job{}
	_ = json.Unmarshal(buf, &job)

	b := buildEntry{
		ID:            0,
		Environment:   []map[string]string{{"SD_TOKEN": "testjwt"}, {"SD_ARTIFACTS_DIR": "/test/artifacts"}, {"SD_API_URL": "http://api-test.screwdriver.cd/v4"}, {"SD_STORE_URL": "http://store-test.screwdriver.cd/v1"}, {"SD_BASE_COMMAND_PATH": "/sd/commands/"}, {"FOO": "foo"}},
		EventID:       0,
		JobID:         0,
		ParentBuildID: []int{0},
		Sha:           "dummy",
		Meta:          Meta{},
		Steps:         job.Steps,
		Image:         job.Image,
		JobName:       "test",
		ArtifactsPath: "sd-artifacts",
	}

	for _, option := range options {
		option(&b)
	}

	return b
}

func TestNew(t *testing.T) {
	t.Run("success with custom artifacts dir", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		_ = json.Unmarshal(buf, &job)
		job.Environment = append(job.Environment, map[string]string{"SD_ARTIFACTS_DIR": "/test/artifacts"})

		config := config.Entry{
			APIURL:   "http://api-test.screwdriver.cd",
			StoreURL: "http://store-test.screwdriver.cd",
			Token:    "testtoken",
			Launcher: config.Launcher{Version: "latest", Image: "screwdrivercd/launcher"},
		}

		expectedBuildEntry := newBuildEntry()
		expectedBuildEntry.Environment[1] = map[string]string{"SD_ARTIFACTS_DIR": "/sd/workspace/artifacts"}
		expectedBuildEntry.Environment = append(expectedBuildEntry.Environment, map[string]string{"SD_ARTIFACTS_DIR": "/test/artifacts"})
		expectedBuildEntry.SrcPath = "/test/sd-local/build/repo"

		option := Option{
			Job:           job,
			Entry:         config,
			JobName:       "test",
			JWT:           "testjwt",
			ArtifactsPath: "sd-artifacts",
			SrcPath:       "/test/sd-local/build/repo",
			Meta:          Meta{},
		}

		launcher := New(option)
		l, ok := launcher.(*launch)
		assert.True(t, ok)
		assert.Equal(t, expectedBuildEntry, l.buildEntry)
	})

	t.Run("success with default artifacts dir", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		_ = json.Unmarshal(buf, &job)

		config := config.Entry{
			APIURL:   "http://api-test.screwdriver.cd",
			StoreURL: "http://store-test.screwdriver.cd",
			Token:    "testtoken",
			Launcher: config.Launcher{Version: "latest", Image: "screwdrivercd/launcher"},
		}

		expectedBuildEntry := newBuildEntry()
		// expectedBuildEntry.Environment[1] corresponds to SD_ARTIFACTS_DIR
		expectedBuildEntry.Environment[1] = map[string]string{"SD_ARTIFACTS_DIR": "/sd/workspace/artifacts"}

		option := Option{
			Job:           job,
			Entry:         config,
			JobName:       "test",
			JWT:           "testjwt",
			ArtifactsPath: "sd-artifacts",
			Meta:          Meta{},
		}

		launcher := New(option)
		l, ok := launcher.(*launch)
		assert.True(t, ok)
		assert.Equal(t, expectedBuildEntry, l.buildEntry)
	})
}

type mockRunner struct {
	errorRunBuild    error
	errorSetupBin    error
	killCalledCount  int
	cleanCalledCount int
}

func (m *mockRunner) runBuild(buildEntry buildEntry) error {
	return m.errorRunBuild
}

func (m *mockRunner) setupBin() error {
	return m.errorSetupBin
}

func (m *mockRunner) clean() {
	m.cleanCalledCount++
}

func (m *mockRunner) kill(os.Signal) {
	m.killCalledCount++
}

func TestRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		_ = json.Unmarshal(buf, &job)

		launch := launch{
			buildEntry: newBuildEntry(),
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
		_ = json.Unmarshal(buf, &job)

		launch := launch{
			buildEntry: newBuildEntry(),
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
		_ = json.Unmarshal(buf, &job)

		launch := launch{
			buildEntry: newBuildEntry(),
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
		_ = json.Unmarshal(buf, &job)

		launch := launch{
			buildEntry: newBuildEntry(),
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

func TestKill(t *testing.T) {
	t.Run("success to call kill", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		_ = json.Unmarshal(buf, &job)

		launch := launch{
			buildEntry: newBuildEntry(),
			runner: &mockRunner{
				errorRunBuild: nil,
				errorSetupBin: nil,
			},
		}
		launch.Kill(syscall.SIGINT)
		mRunner := launch.runner.(*mockRunner)
		assert.Equal(t, 1, mRunner.killCalledCount)
	})
}

func TestClean(t *testing.T) {
	t.Run("success to call clean", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))
		job := screwdriver.Job{}
		_ = json.Unmarshal(buf, &job)

		launch := launch{
			buildEntry: newBuildEntry(),
			runner: &mockRunner{
				errorRunBuild: nil,
				errorSetupBin: nil,
			},
		}
		launch.Clean()
		mRunner := launch.runner.(*mockRunner)
		assert.Equal(t, 1, mRunner.cleanCalledCount)
	})
}
