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
	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

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

		expected1 := []map[string]string{{"hoge": "fuga"}, {"foo": "bar"}}
		expected2 := []map[string]string{{"foo": "bar"}, {"hoge": "fuga"}}

		launchNew = func(option launch.Option) launch.Launcher {
			assert.Contains(t, [...]screwdriver.EnvVars{expected1, expected2}, option.OptionEnv)
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

		expected1 := []map[string]string{{"hoge": "fuga"}, {"foo": "bar"}}
		expected2 := []map[string]string{{"foo": "bar"}, {"hoge": "fuga"}}

		launchNew = func(option launch.Option) launch.Launcher {
			assert.Contains(t, [...]screwdriver.EnvVars{expected1, expected2}, option.OptionEnv)
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

		expected1 := []map[string]string{{"hoge": "fuga"}, {"foo": "bar"}, {"hoge": "overwritten"}, {"baz": "qux"}}
		expected2 := []map[string]string{{"hoge": "fuga"}, {"foo": "bar"}, {"baz": "qux"}, {"hoge": "overwritten"}}
		expected3 := []map[string]string{{"foo": "bar"}, {"hoge": "fuga"}, {"hoge": "overwritten"}, {"baz": "qux"}}
		expected4 := []map[string]string{{"foo": "bar"}, {"hoge": "fuga"}, {"baz": "qux"}, {"hoge": "overwritten"}}

		launchNew = func(option launch.Option) launch.Launcher {
			assert.Contains(t, [...]screwdriver.EnvVars{expected1, expected2, expected3, expected4}, option.OptionEnv)
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
		wantErr := "can't pass the both options `meta` and `meta-file`, please specify only one of them"
		want := "Usage:\n  build [job name] [flags]\n" +
			buildLocalFlags()
		assert.Equal(t, wantErr, err.Error())
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})

	t.Run("Failed build cmd when too many args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{"test", "main"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		wantErr := "accepts 1 arg(s), received 2"
		want := "Usage:\n  build [job name] [flags]\n" +
			buildLocalFlags()
		assert.Equal(t, wantErr, err.Error())
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})

	t.Run("Failed build cmd when too little args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{})

		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		wantErr := "accepts 1 arg(s), received 0"
		want := "Usage:\n  build [job name] [flags]\n" +
			buildLocalFlags()
		assert.Equal(t, wantErr, err.Error())
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
