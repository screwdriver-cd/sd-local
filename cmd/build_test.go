package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/screwdriver-cd/sd-local/launch"
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
  -m, --memory string          Memory limit for build container, which take a positive integer, followed by a suffix of b, k, m, g.
      --meta string            Metadata to pass into the build environment, which is represented with JSON format
      --meta-file string       Path to the meta file. meta file is represented with JSON format.
      --src-url string         Specify the source url to build.
                               ex) git@github.com:<org>/<repo>.git[#<branch>]
                                   https://github.com/<org>/<repo>.git[#<branch>]
      --sudo                   Use sudo command for container runtime.

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
}
