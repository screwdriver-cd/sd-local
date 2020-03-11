package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCmd(t *testing.T) {
	t.Run("Success version cmd", func(t *testing.T) {
		cmd := newVersionCmd()
		buf := bytes.NewBuffer(nil)
		cmd.SetOut(buf)
		cmd.Execute()
		assert.Equal(t, "dev\n", buf.String())
	})
}
