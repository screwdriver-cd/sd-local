package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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

	t.Run("Failed build cmd when too many args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{"test", "main"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := `Error: accepts 1 arg(s), received 2
Usage:
  build [job name] [flags]

Flags:
      --artifacts-dir string   Path to the host side directory which is mounted into $SD_ARTIFACTS_DIR. (default "sd-artifacts")
  -h, --help                   help for build

`
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})

	t.Run("Failed build cmd when too little args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{})

		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := `Error: accepts 1 arg(s), received 0
Usage:
  build [job name] [flags]

Flags:
      --artifacts-dir string   Path to the host side directory which is mounted into $SD_ARTIFACTS_DIR. (default "sd-artifacts")
  -h, --help                   help for build

`
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})
}
