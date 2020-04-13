package config

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-yaml/yaml"

	"github.com/stretchr/testify/assert"
)

var testDir string = "./testdata"

func TestCreateConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		cnfPath := filepath.Join(testDir, fmt.Sprintf("%vconfig", rand.Int()))
		defer os.Remove(cnfPath)

		expect := configList{
			Configs: map[string]Config{
				"default": {
					APIURL:   "",
					StoreURL: "",
					Token:    "",
					Launcher: Launcher{
						Version: "stable",
						Image:   "screwdrivercd/launcher",
					},
				},
			},
			Current: "default",
		}

		err := create(cnfPath)
		assert.Nil(t, err)
		file, _ := os.Open(cnfPath)
		actual := configList{}
		_ = yaml.NewDecoder(file).Decode(&actual)
		assert.Equal(t, expect, actual)
	})

	t.Run("success by exists file", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		cnfPath := filepath.Join(testDir, fmt.Sprintf("%vconfig", rand.Int()))
		defer os.Remove(cnfPath)

		expect := configList{
			Configs: map[string]Config{
				"default": {
					APIURL:   "",
					StoreURL: "",
					Token:    "",
					Launcher: Launcher{
						Version: "stable",
						Image:   "screwdrivercd/launcher",
					},
				},
			},
			Current: "default",
		}

		err := create(cnfPath)
		assert.Nil(t, err)
		err = create(cnfPath)
		assert.Nil(t, err)

		file, _ := os.Open(cnfPath)
		actual := configList{}
		_ = yaml.NewDecoder(file).Decode(&actual)
		assert.Equal(t, expect, actual)
	})
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cnfPath := filepath.Join(testDir, "successConfig")

		testConfig := Config{
			APIURL:   "api-url",
			StoreURL: "store-api-url",
			Token:    "dummy_token",
			Launcher: Launcher{
				Version: "latest",
				Image:   "screwdrivercd/launcher",
			},
			filePath: cnfPath,
		}

		actual, err := New(cnfPath)

		assert.Nil(t, err)
		assert.Equal(t, cnfPath, actual.filePath)
		assert.Equal(t, testConfig, actual)
	})

	t.Run("failure by invalid current", func(t *testing.T) {
		cnfPath := filepath.Join(testDir, "failureCurrentConfig")
		_, err := New(cnfPath)

		assert.Equal(t, "config `doesnotexist` does not exist", err.Error())
	})

	t.Run("failure by invalid yaml", func(t *testing.T) {
		cnfPath := filepath.Join(testDir, "failureConfig")
		_, err := New(cnfPath)

		assert.Equal(t, 0, strings.Index(err.Error(), "failed to parse config file: "), fmt.Sprintf("expected error is `failed to parse config file: ...`, actual: `%v`", err))
	})
}

func TestSetConfig(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	cnfPath := filepath.Join(testDir, ".sdlocal", fmt.Sprintf("%vconfig", rand.Int()))
	defer os.RemoveAll(filepath.Join(testDir, ".sdlocal"))

	testCases := []struct {
		name       string
		setting    map[string]string
		expectConf Config
	}{
		{

			name: "success",
			setting: map[string]string{
				"api-url":          "example-api.com",
				"store-url":        "example-store.com",
				"token":            "dummy-token",
				"launcher-version": "1.0.0",
				"launcher-image":   "alpine",
				"invalidKey":       "invalidValue",
			},
			expectConf: Config{
				APIURL:   "example-api.com",
				StoreURL: "example-store.com",
				Token:    "dummy-token",
				Launcher: Launcher{
					Version: "1.0.0",
					Image:   "alpine",
				},
				filePath: cnfPath,
			},
		},
		{
			name: "success override",
			setting: map[string]string{
				"api-url":          "override-example-api.com",
				"store-url":        "override-example-store.com",
				"token":            "override-dummy-token",
				"launcher-version": "override-1.0.0",
				"launcher-image":   "override-alpine",
				"invalidKey":       "override-invalidValue",
			},
			expectConf: Config{
				APIURL:   "override-example-api.com",
				StoreURL: "override-example-store.com",
				Token:    "override-dummy-token",
				Launcher: Launcher{
					Version: "override-1.0.0",
					Image:   "override-alpine",
				},
				filePath: cnfPath,
			},
		},
		{
			name: "used default value",
			setting: map[string]string{
				"api-url":          "override-example-api.com",
				"store-url":        "override-example-store.com",
				"token":            "override-dummy-token",
				"launcher-version": "",
				"launcher-image":   "",
				"invalidKey":       "override-invalidValue",
			},
			expectConf: Config{
				APIURL:   "override-example-api.com",
				StoreURL: "override-example-store.com",
				Token:    "override-dummy-token",
				Launcher: Launcher{
					Version: "stable",
					Image:   "screwdrivercd/launcher",
				},
				filePath: cnfPath,
			},
		},
		{
			name: "invalid key",
			setting: map[string]string{
				"api-url":          "override-example-api.com",
				"store-url":        "override-example-store.com",
				"token":            "override-dummy-token",
				"launcher-version": "",
				"launcher-image":   "",
				"invalidKey":       "invalidValue",
			},
			expectConf: Config{
				APIURL:   "override-example-api.com",
				StoreURL: "override-example-store.com",
				Token:    "override-dummy-token",
				Launcher: Launcher{
					Version: "stable",
					Image:   "screwdrivercd/launcher",
				},
				filePath: cnfPath,
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := New(cnfPath)
			for key, val := range tt.setting {
				err := c.Set(key, val)
				if key == "invalidKey" {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}
			assert.Equal(t, tt.expectConf, c)

			_, err := os.Stat(cnfPath)
			assert.Nil(t, err)

			actual, _ := New(cnfPath)
			assert.Equal(t, tt.expectConf, actual)
		})
	}
}
