package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewCmd(t *testing.T) {
	fp := filePath
	defer func() {
		filePath = fp
	}()

	filePath = func(isLocal bool) (string, error) {
		if isLocal {
			return "./testdata/local_config", nil
		}
		return "./testdata/config", nil
	}

	testCase := []struct {
		name   string
		args   []string
		expect string
	}{
		{
			name: "success by not use local config",
			args: []string{"view"},
			expect: `KEY               VALUE
api-url           api.screwdriver.com
store-url         store.screwdriver.com
token             sd-token
launcher-version  1.0.0
launcher-image    screwdrivercd/launcher
`,
		},
		{
			name: "success by use local config",
			args: []string{"view", "--local"},
			expect: `KEY               VALUE
api-url           local.api.screwdriver.com
store-url         local.store.screwdriver.com
token             local.sd-token
launcher-version  stable
launcher-image    screwdrivercd/launcher
`,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewConfigCmd()
			cmd.SetArgs(tt.args)
			buf := bytes.NewBuffer(nil)
			cmd.SetOut(buf)
			cmd.Execute()
			assert.Equal(t, tt.expect, buf.String())
		})
	}
}
