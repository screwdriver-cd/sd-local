package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/screwdriver-cd/sd-local/buildlog"
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"

	"github.com/screwdriver-cd/sd-local/screwdriver"
)

type mockAPI struct{}
type mockLogger struct{}
type mockLaunch struct{}

func (mock mockAPI) Job(jobName, filePath string) (screwdriver.Job, error) {
	return screwdriver.Job{}, nil
}

func (mock mockAPI) JWT() string { return "" }

func (mock mockAPI) InitJWT() error { return nil }

func (mock mockLogger) Run() {}

func (mock mockLogger) Stop() { close(loggerDone) }

func (mock mockLaunch) Run() error { return nil }

func (mock mockLaunch) Kill(os.Signal) {}

func (mock mockLaunch) Clean() {}

func setup() {
	configNew = func(confPath string) (config.Config, error) {
		defaultEntry := &config.Entry{
			Launcher: config.Launcher{
				Version: "stable",
				Image:   "screwdrivercd/launcher",
			},
		}

		return config.Config{
			Entries: map[string]*config.Entry{
				"default": defaultEntry,
			},
			Current: "default",
		}, nil
	}
	apiNew = func(url, token string) screwdriver.API { return mockAPI{} }
	buildLogNew = func(filepath string, writer io.Writer, done chan<- struct{}) (logger buildlog.Logger, err error) {
		return mockLogger{}, nil
	}
	launchNew = func(option launch.Option) launch.Launcher {
		return mockLaunch{}
	}
	osMkdirAll = func(path string, filemode os.FileMode) error { return nil }
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	os.Exit(ret)
}

func TestRootCmd(t *testing.T) {
	t.Run("Success root cmd", func(t *testing.T) {
		root := newRootCmd()
		root.AddCommand(newBuildCmd())
		root.SetArgs([]string{})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := "Run build instantly on your local machine with\na mostly the same environment as Screwdriver.cd's\n\nUsage:\n  sd-local [command]\n\nAvailable Commands:\n  build       Run screwdriver build.\n  help        Help about any command\n\nFlags:\n  -h, --help      help for sd-local\n  -v, --verbose   verbose output.\n\nUse \"sd-local [command] --help\" for more information about a command.\n"
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)
	})

	t.Run("Success root cmd displays update", func(t *testing.T) {
		root := newRootCmd()
		root.AddCommand(newUpdateCmd())
		root.SetArgs([]string{})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := "Run build instantly on your local machine with\na mostly the same environment as Screwdriver.cd's\n\nUsage:\n  sd-local [command]\n\nAvailable Commands:\n  help        Help about any command\n  update      Update to the latest version\n\nFlags:\n  -h, --help      help for sd-local\n  -v, --verbose   verbose output.\n\nUse \"sd-local [command] --help\" for more information about a command.\n"
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)
	})

	t.Run("Failed root cmd by no arguments for sub command", func(t *testing.T) {
		root := newRootCmd()
		root.AddCommand(newBuildCmd())
		root.SetArgs([]string{"build"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := `Error: accepts 1 arg(s), received 0
Usage:
  sd-local build [job name] [flags]

Flags:
      --artifacts-dir string   Path to the host side directory which is mounted into $SD_ARTIFACTS_DIR. (default "sd-artifacts")
  -e, --env stringToString     Set key and value relationship which is set as environment variables of Build Container. (<key>=<value>) (default [])
      --env-file string        Path to config file of environment variables. '.env' format file can be used.
  -h, --help                   help for build
  -i, --interactive            Attach the build container in interactive mode.
  -m, --memory string          Memory limit for build container, which take a positive integer, followed by a suffix of b, k, m, g.
      --meta string            Metadata to pass into the build environment, which is represented with JSON format
      --meta-file string       Path to the meta file. meta file is represented with JSON format.
      --privileged             Use privileged mode for container runtime.
      --src-url string         Specify the source url to build.
                               ex) git@github.com:<org>/<repo>.git[#<branch>]
                                   https://github.com/<org>/<repo>.git[#<branch>]
      --sudo                   Use sudo command for container runtime.

Global Flags:
  -v, --verbose   verbose output.

`
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})

	t.Run("Failed root cmd by invalid sub command", func(t *testing.T) {
		root := newRootCmd()
		root.AddCommand(newBuildCmd())
		root.SetArgs([]string{"hoge"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := "Error: unknown command \"hoge\" for \"sd-local\"\nRun 'sd-local --help' for usage.\n"
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})
}
