package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type a struct {
	a string
}

func TestBuildCmd(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	root := newRootCmd()
	root.SetArgs([]string{"build", "main"})
	root.SetOutput(buf)
	err := root.Execute()
	assert.Nil(t, err)
}
