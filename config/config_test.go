package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDir string = "./testdata"

func TestReadConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testConfig := Config{
			ApiUrl:   "api-url",
			StoreUrl: "store-api-url",
			Token:    "dummy_token",
			Launcher: Launcher{
				Version: "latest",
				Image:   "screwdrivercd/launcher",
			},
		}

		actual, err := ReadConfig(testDir)

		assert.Nil(t, err)
		assert.Equal(t, testConfig, actual)
	})

	t.Run("failure by no entry", func(t *testing.T) {
		emptyConfig := Config{}
		actual, err := ReadConfig("./not-exist")
		assert.NotNil(t, err, fmt.Sprintf("There is no error when reading config fails"))

		assert.Equal(t, emptyConfig, actual)

		msg := err.Error()
		assert.Equal(t, 0, strings.Index(msg, "failed to read config file: "), fmt.Sprintf("expected error is `failed to read config file: ...`, actual: `%v`", msg))
	})

	t.Run("failure by invalid yaml", func(t *testing.T) {
		invalidConfigFile := []byte("key; 1")
		tmpdir, err := ioutil.TempDir("", "sdlocal")

		assert.Nil(t, err, fmt.Sprintf("failed to create temporary directory: %v", err))

		defer os.RemoveAll(tmpdir)

		err = ioutil.WriteFile(path.Join(tmpdir, "config"), invalidConfigFile, 0666)

		assert.Nil(t, err, fmt.Sprintf("failed to create config file: %v", err))

		emptyConfig := Config{}
		actual, err := ReadConfig(tmpdir)

		assert.Equal(t, emptyConfig, actual)

		msg := fmt.Errorf("%v", err).Error()
		assert.Equal(t, 0, strings.Index(msg, "failed to parse config file: "), fmt.Sprintf("expected error is `failed to parse config file: ...`, actual: `%v`", msg))
	})
}
