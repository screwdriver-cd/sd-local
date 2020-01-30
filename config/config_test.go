package config

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDir string = "./testdata"

func TestReadConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testConfig := Config{
			APIURL:   "api-url",
			StoreURL: "store-api-url",
			Token:    "dummy_token",
			Launcher: Launcher{
				Version: "latest",
				Image:   "screwdrivercd/launcher",
			},
		}

		cnfPath := path.Join(testDir, "successConfig")
		actual, err := ReadConfig(cnfPath)

		assert.Nil(t, err)
		assert.Equal(t, testConfig, actual)
	})

	t.Run("failure by no entry", func(t *testing.T) {
		emptyConfig := Config{}
		actual, err := ReadConfig("./not-exist")
		assert.NotNil(t, err, fmt.Sprintf("There is no error when reading config files"))

		assert.Equal(t, emptyConfig, actual)

		msg := err.Error()
		assert.Equal(t, 0, strings.Index(msg, "failed to read config file: "), fmt.Sprintf("expected error is `failed to read config file: ...`, actual: `%v`", msg))
	})

	t.Run("failure by invalid yaml", func(t *testing.T) {
		cnfPath := path.Join(testDir, "failureConfig")
		_, err := ReadConfig(cnfPath)

		assert.Equal(t, 0, strings.Index(err.Error(), "failed to parse config file: "), fmt.Sprintf("expected error is `failed to parse config file: ...`, actual: `%v`", err))
	})
}
