package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/screwdriver-cd/sd-local/config"
)

func Example_newConfigViewCmd() {
	cmd := NewConfigCmd()
	cmd.SetArgs([]string{"view"})
	cmd.SetOut(os.Stdout)

	c := config.Config{
		APIURL:   "api.screwdriver.com",
		StoreURL: "store.screwdriver.com",
		Token:    "token",
		Launcher: config.Launcher{
			Version: "latest",
			Image:   "screwdrivercd/launcher",
		},
	}
	configNew = func(configPath string) (config.Config, error) {
		return c, nil
	}

	cmd.Execute()
	// Output:
	// KEY               VALUE
	// api-url           api.screwdriver.com
	// store-url         store.screwdriver.com
	// token             token
	// launcher-version  latest
	// launcher-image    screwdrivercd/launcher
}

func Example_newConfigViewCmd_local() {
	cmd := NewConfigCmd()
	cmd.SetArgs([]string{"view", "--local"})
	// buf := bytes.NewBuffer(nil)
	cmd.SetOut(os.Stdout)
	cmd.Execute()
	// Output:
	// KEY               VALUE
	// api-url           local.api.screwdriver.com
	// store-url         local.store.screwdriver.com
	// token             local.token
	// launcher-version  latest
	// launcher-image    screwdrivercd/launcher
}

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
