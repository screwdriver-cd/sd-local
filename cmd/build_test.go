package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const buildUsage = `
Usage:
  build [job name] [flags]

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
  -S, --socket string          Path to the socket. It will used in build container.
      --src-url string         Specify the source url to build.
                               ex) git@github.com:<org>/<repo>.git[#<branch>]
                                   https://github.com/<org>/<repo>.git[#<branch>]
      --sudo                   Use sudo command for container runtime.
      --vol strings            Volumes to mount into build container.

`

func TestBuildCmd(t *testing.T) {
	t.Run("Success build cmd", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{"test"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := ""
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)
		assert.Equal(t, "sd-artifacts", artifactsDir)
	})

	t.Run("Success build cmd with --artifacts-dir", func(t *testing.T) {
		defFunc := osMkdirAll
		osMkdirAll = os.MkdirAll
		defer func() {
			osMkdirAll = defFunc
		}()

		root := newBuildCmd()

		dir, err := ioutil.TempDir("", "example")
		if err != nil {
			t.Fatal(err)
		}

		defer os.RemoveAll(dir)

		artifactsDir := filepath.Join(dir, "sd-artifacts")

		root.SetArgs([]string{"test", "--artifacts-dir", artifactsDir})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err = root.Execute()
		want := ""
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)

		_, err = os.Stat(artifactsDir)
		assert.Nil(t, err)
	})

	t.Run("Success build cmd with --env", func(t *testing.T) {
		root := newBuildCmd()

		root.SetArgs([]string{"test", "--env", "hoge=fuga", "-e", "foo=bar"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)

		expected := launch.EnvVar{
			"hoge": "fuga",
			"foo":  "bar",
		}

		launchNew = func(option launch.Option) launch.Launcher {
			assert.Equal(t, expected, option.OptionEnv)
			return mockLaunch{}
		}

		err := root.Execute()
		want := ""
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)
		assert.Equal(t, "sd-artifacts", artifactsDir)
	})

	t.Run("Success build cmd with --env-file", func(t *testing.T) {
		root := newBuildCmd()

		root.SetArgs([]string{"test", "--env-file", "./testdata/test_env"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)

		expected := launch.EnvVar{
			"hoge": "fuga",
			"foo":  "bar",
		}

		launchNew = func(option launch.Option) launch.Launcher {
			assert.Equal(t, expected, option.OptionEnv)
			return mockLaunch{}
		}

		err := root.Execute()
		want := ""
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)
		assert.Equal(t, "sd-artifacts", artifactsDir)
	})

	t.Run("Success build cmd with --env and --env-file", func(t *testing.T) {
		root := newBuildCmd()

		root.SetArgs([]string{"test", "--env-file", "./testdata/test_env", "--env", "hoge=overwritten", "-e", "baz=qux"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)

		expected := launch.EnvVar{
			"hoge": "overwritten",
			"foo":  "bar",
			"baz":  "qux",
		}

		launchNew = func(option launch.Option) launch.Launcher {
			assert.Equal(t, expected, option.OptionEnv)
			return mockLaunch{}
		}

		err := root.Execute()
		want := ""
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)
		assert.Equal(t, "sd-artifacts", artifactsDir)
	})

	t.Run("Success build cmd with --meta", func(t *testing.T) {
		root := newBuildCmd()

		root.SetArgs([]string{"test", "--meta", "{\"hoge\":\"fuga\"}"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)

		expected := launch.Meta{
			"hoge": "fuga",
		}

		launchNew = func(option launch.Option) launch.Launcher {
			assert.Equal(t, expected, option.Meta)
			return mockLaunch{}
		}

		err := root.Execute()
		want := ""
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)
	})

	t.Run("Success build cmd with --meta-file", func(t *testing.T) {
		root := newBuildCmd()

		root.SetArgs([]string{"test", "--meta-file", "./testdata/test_meta.json"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)

		expected := launch.Meta{
			"hoge": "fuga",
			"foo": map[string]interface{}{
				"bar": "aaa",
			},
		}

		launchNew = func(option launch.Option) launch.Launcher {
			assert.Equal(t, expected, option.Meta)
			return mockLaunch{}
		}

		err := root.Execute()
		want := ""
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)
	})

	t.Run("Failed build cmd with --meta and --meta-file", func(t *testing.T) {
		root := newBuildCmd()

		root.SetArgs([]string{"test", "--meta-file", "./testdata/test_meta.json", "--meta", "{\"hoge\":\"fuga\"}"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)

		launchNew = func(option launch.Option) launch.Launcher {
			return mockLaunch{}
		}

		err := root.Execute()
		want := "Error: can't pass the both options `meta` and `meta-file`, please specify only one of them" + buildUsage
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})

	t.Run("Failed build cmd when too many args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{"test", "main"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := "Error: accepts 1 arg(s), received 2" + buildUsage
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})

	t.Run("Failed build cmd when too little args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{})

		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := "Error: accepts 1 arg(s), received 0" + buildUsage
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})

	t.Run("Output y/n message on build cmd without User-Agent", func(t *testing.T) {
		root := newBuildCmd()

		root.SetArgs([]string{"test"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)

		logBuf := bytes.NewBuffer(nil)
		logrus.SetOutput(logBuf)

		configNew = func(confPath string) (config.Config, error) {
			defaultEntry := &config.Entry{
				Launcher: config.Launcher{
					Version: "stable",
					Image:   "screwdrivercd/launcher",
				},
				UUID: "",
			}

			return config.Config{
				Entries: map[string]*config.Entry{
					"default": defaultEntry,
				},
				Current: "default",
			}, nil
		}

		actualOutput := captureStdout(func() {
			root.Execute()
		})

		expectOutput := `sd-local collects UUIDs for statistical surveys.
You can reset it later by removing the UUID key from config.
Would you please cooperate with the survey? [y/N]: `
		assert.Equal(t, expectOutput, actualOutput)
	})

	t.Run("Not output y/n message on build cmd with User-Agent", func(t *testing.T) {
		root := newBuildCmd()

		root.SetArgs([]string{"test"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)

		logBuf := bytes.NewBuffer(nil)
		logrus.SetOutput(logBuf)

		configNew = func(confPath string) (config.Config, error) {
			defaultEntry := &config.Entry{
				Launcher: config.Launcher{
					Version: "stable",
					Image:   "screwdrivercd/launcher",
				},
				UUID: "foo",
			}

			return config.Config{
				Entries: map[string]*config.Entry{
					"default": defaultEntry,
				},
				Current: "default",
			}, nil
		}

		actualOutput := captureStdout(func() {
			root.Execute()
		})

		expectOutput := ""
		assert.Equal(t, expectOutput, actualOutput)
	})
}

func captureStdout(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	stdout := os.Stdout
	os.Stdout = w

	f()

	os.Stdout = stdout
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}
