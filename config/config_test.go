package config

import (
	"fmt"
	"io/ioutil"
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

		expect := ConfigList{
			Configs: map[string]*Config{
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
		actual := ConfigList{}
		_ = yaml.NewDecoder(file).Decode(&actual)
		assert.Equal(t, expect, actual)
	})

	t.Run("success by exists file", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		cnfPath := filepath.Join(testDir, fmt.Sprintf("%vconfig", rand.Int()))
		defer os.Remove(cnfPath)

		expect := ConfigList{
			Configs: map[string]*Config{
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
		actual := ConfigList{}
		_ = yaml.NewDecoder(file).Decode(&actual)
		assert.Equal(t, expect, actual)
	})
}

func TestNewConfigList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cnfPath := filepath.Join(testDir, "successConfig")

		actual, err := NewConfigList(cnfPath)
		if err != nil {
			t.Fatal(err)
		}

		testConfigList := ConfigList{
			Configs: map[string]*Config{
				"default": {
					APIURL:   "api-url",
					StoreURL: "store-api-url",
					Token:    "dummy_token",
					Launcher: Launcher{
						Version: "latest",
						Image:   "screwdrivercd/launcher",
					},
				},
			},
			Current:  "default",
			filePath: cnfPath,
		}

		assert.Nil(t, err)
		assert.Equal(t, testConfigList, actual)
	})

	t.Run("failure by invalid yaml", func(t *testing.T) {
		cnfPath := filepath.Join(testDir, "failureConfig")

		_, err := NewConfigList(cnfPath)

		assert.Equal(t, 0, strings.Index(err.Error(), "failed to parse config file: "), fmt.Sprintf("expected error is `failed to parse config file: ...`, actual: `%v`", err))
	})
}

func TestConfigListGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		configList := ConfigList{
			Configs: map[string]*Config{
				"default": {
					APIURL:   "api-url",
					StoreURL: "store-api-url",
					Token:    "dummy_token",
					Launcher: Launcher{
						Version: "latest",
						Image:   "screwdrivercd/launcher",
					},
				},
			},
			Current: "default",
		}

		testConfig := &Config{
			APIURL:   "api-url",
			StoreURL: "store-api-url",
			Token:    "dummy_token",
			Launcher: Launcher{
				Version: "latest",
				Image:   "screwdrivercd/launcher",
			},
		}

		actual, err := configList.Get("default")
		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, err)
		assert.Equal(t, testConfig, actual)
	})

	t.Run("failure by invalid current", func(t *testing.T) {
		configList := ConfigList{
			Configs: map[string]*Config{
				"default": {
					APIURL:   "api-url",
					StoreURL: "store-api-url",
					Token:    "dummy_token",
					Launcher: Launcher{
						Version: "latest",
						Image:   "screwdrivercd/launcher",
					},
				},
			},
			Current: "doesnotexist",
		}

		_, err := configList.Get(configList.Current)

		assert.Equal(t, "config `doesnotexist` does not exist", err.Error())
	})
}

func TestConfigListAdd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		configList := ConfigList{
			Current: "default",
			Configs: map[string]*Config{
				"default": {
					APIURL:   "api-url",
					StoreURL: "store-api-url",
					Token:    "dummy_token",
					Launcher: Launcher{
						Version: "latest",
						Image:   "screwdrivercd/launcher",
					},
				},
			},
		}

		err := configList.Add("test")
		if err != nil {
			t.Fatal(err)
		}

		expected := ConfigList{
			Configs: map[string]*Config{
				"default": {
					APIURL:   "api-url",
					StoreURL: "store-api-url",
					Token:    "dummy_token",
					Launcher: Launcher{
						Version: "latest",
						Image:   "screwdrivercd/launcher",
					},
				},
				"test": {
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

		assert.Equal(t, expected, configList)
	})

	t.Run("failure by the name that exists", func(t *testing.T) {
		configList := ConfigList{
			Current: "default",
			Configs: map[string]*Config{
				"default": {
					APIURL:   "api-url",
					StoreURL: "store-api-url",
					Token:    "dummy_token",
					Launcher: Launcher{
						Version: "latest",
						Image:   "screwdrivercd/launcher",
					},
				},
			},
		}

		err := configList.Add("default")
		assert.Equal(t, "config `default` already exists", err.Error())
	})
}

func TestConfigListSave(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		cnfPath := filepath.Join(testDir, ".sdlocal", fmt.Sprintf("%vconfig", rand.Int()))
		err := os.MkdirAll(filepath.Join(testDir, ".sdlocal"), 0777)
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(filepath.Join(testDir, ".sdlocal"))

		file, err := os.Create(cnfPath)
		if err != nil {
			t.Fatal(err)
		}
		file.Close()

		configList := ConfigList{
			Current: "default",
			Configs: map[string]*Config{
				"default": {
					APIURL:   "api-url",
					StoreURL: "store-api-url",
					Token:    "dummy_token",
					Launcher: Launcher{
						Version: "latest",
						Image:   "screwdrivercd/launcher",
					},
				},
			},
			filePath: cnfPath,
		}

		err = configList.Save()
		if err != nil {
			t.Fatal(err)
		}

		actual, err := ioutil.ReadFile(cnfPath)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := yaml.Marshal(configList)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expected, actual)
	})
}

func TestSetConfig(t *testing.T) {
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
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			c := &Config{}

			for key, val := range tt.setting {
				err := c.Set(key, val)
				if key == "invalidKey" {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}
			assert.Equal(t, tt.expectConf, *c)
		})
	}
}
