package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/screwdriver-cd/sd-local/config"
	"github.com/stretchr/testify/assert"
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

	cmd := NewConfigCmd()
	cmd.SetArgs([]string{"view"})
	buf := bytes.NewBuffer(nil)
	cmd.SetOut(buf)
	// c := config.Config{
	// 	APIURL:   "api.screwdriver.com",
	// 	StoreURL: "store.screwdriver.com",
	// 	Token:    "sd-token",
	// 	Launcher: config.Launcher{
	// 		Version: "1.0.0",
	// 		Image:   "screwdrivercd/launcher",
	// 	},
	// }
	filePath = func(isLocal bool) (string, error) {
		if isLocal {
			return "./testdata/local_config", nil
		}
		return "./testdata/config", nil
	}

	cmd.Execute()

	expect := `KEY               VALUE
api-url           api.screwdriver.com
store-url         store.screwdriver.com
token             sd-token
launcher-version  1.0.0
launcher-image    screwdrivercd/launcher
`

	assert.Equal(t, expect, buf.String())
}
