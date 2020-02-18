package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCmd(t *testing.T) {
	setupTest()

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

	t.Run("Failed build cmd when to many args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{"test", "main"})
		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := ""
		assert.NotEqual(t, want, buf.String())
		assert.NotNil(t, err)
	})

	t.Run("Failed build cmd when to little args", func(t *testing.T) {
		root := newBuildCmd()
		root.SetArgs([]string{})

		buf := bytes.NewBuffer(nil)
		root.SetOut(buf)
		err := root.Execute()
		want := ""
		assert.NotEqual(t, want, buf.String())
		assert.NotNil(t, err)
	})
}
