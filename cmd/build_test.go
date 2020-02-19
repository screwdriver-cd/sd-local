package cmd

import (
	"bytes"
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
	})

	t.Run("Failed build cmd when too many args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{"test", "main"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := "Error: accepts 1 arg(s), received 2\nUsage:\n  build [job name] [flags]\n\nFlags:\n  -h, --help   help for build\n\n"
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})

	t.Run("Failed build cmd when too little args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{})

		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := "Error: accepts 1 arg(s), received 0\nUsage:\n  build [job name] [flags]\n\nFlags:\n  -h, --help   help for build\n\n"
		assert.Equal(t, want, buf.String())
		assert.NotNil(t, err)
	})
}
