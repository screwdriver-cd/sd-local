package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCmd(t *testing.T) {
	cases := []struct {
		name    string
		version string
		expect  string
	}{
		{
			name:    "0.0.1 is embedded as version",
			version: "0.0.1",
			expect:  "0.0.1",
		},
		{
			name:    "0.1.0 is embedded as version",
			version: "0.1.0",
			expect:  "0.1.0",
		},
		{
			name:   "version is not embedded",
			expect: "dev",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tv := version
			if c.version != "" {
				version = c.version
			}
			defer func() { version = tv }()
			cmd := newVersionCmd()
			buf := bytes.NewBuffer(nil)
			cmd.SetOut(buf)
			cmd.Execute()
			assert.Equal(t, fmt.Sprintf("%s\n", c.expect), buf.String())
		})
	}
}
