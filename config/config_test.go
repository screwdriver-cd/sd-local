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

		expect := Config{
			Entries: map[string]*Entry{
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
		actual := Config{}
		_ = yaml.NewDecoder(file).Decode(&actual)
		assert.Equal(t, expect, actual)
	})

	t.Run("success by exists file", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		cnfPath := filepath.Join(testDir, fmt.Sprintf("%vconfig", rand.Int()))
		defer os.Remove(cnfPath)

		expect := Config{
			Entries: map[string]*Entry{
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
		actual := Config{}
		_ = yaml.NewDecoder(file).Decode(&actual)
		assert.Equal(t, expect, actual)
	})
}
func TestNewConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cnfPath := filepath.Join(testDir, "successConfig")

		actual, err := New(cnfPath)
		if err != nil {
			t.Fatal(err)
		}

		testConfig := Config{
			Entries: map[string]*Entry{
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
		assert.Equal(t, testConfig, actual)
	})

	t.Run("failure by invalid yaml", func(t *testing.T) {
		cnfPath := filepath.Join(testDir, "failureConfig")

		_, err := New(cnfPath)

		assert.Equal(t, 0, strings.Index(err.Error(), "failed to parse config file: "), fmt.Sprintf("expected error is `failed to parse config file: ...`, actual: `%v`", err))
	})
}

func TestConfigEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		config := Config{
			Entries: map[string]*Entry{
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

		testEntry := &Entry{
			APIURL:   "api-url",
			StoreURL: "store-api-url",
			Token:    "dummy_token",
			Launcher: Launcher{
				Version: "latest",
				Image:   "screwdrivercd/launcher",
			},
		}

		actual, err := config.Entry("default")
		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, err)
		assert.Equal(t, testEntry, actual)
	})

	t.Run("failure by invalid current", func(t *testing.T) {
		config := Config{
			Entries: map[string]*Entry{
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

		_, err := config.Entry(config.Current)

		assert.Equal(t, "config `doesnotexist` does not exist", err.Error())
	})
}

func TestConfigAddEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		config := Config{
			Current: "default",
			Entries: map[string]*Entry{
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

		err := config.AddEntry("test")
		if err != nil {
			t.Fatal(err)
		}

		expected := Config{
			Entries: map[string]*Entry{
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

		assert.Equal(t, expected, config)
	})

	t.Run("failure by the name that exists", func(t *testing.T) {
		config := Config{
			Current: "default",
			Entries: map[string]*Entry{
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

		err := config.AddEntry("default")
		assert.Equal(t, "config `default` already exists", err.Error())
	})
}

func TestConfigDeleteEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		config := Config{
			Current: "default",
			Entries: map[string]*Entry{
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

		err := config.DeleteEntry("test")
		if err != nil {
			t.Fatal(err)
		}

		expected := Config{
			Current: "default",
			Entries: map[string]*Entry{
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

		assert.Equal(t, expected, config)
	})

	t.Run("failure", func(t *testing.T) {
		config := Config{
			Current: "default",
			Entries: map[string]*Entry{
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

		err := config.DeleteEntry("test")
		assert.Equal(t, "config `test` does not exist", err.Error())
	})

	t.Run("failure by trying to delete current entry", func(t *testing.T) {
		config := Config{
			Current: "default",
			Entries: map[string]*Entry{
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

		err := config.DeleteEntry("default")
		assert.Equal(t, "config `default` is current config", err.Error())
	})
}

func TestConfigSetCurrent(t *testing.T) {
	testConfig := func() Config {
		return Config{
			Current: "default",
			Entries: map[string]*Entry{
				"default": {
					APIURL:   "api-url",
					StoreURL: "store-api-url",
					Token:    "dummy_token",
					Launcher: Launcher{},
				},
				"test": {
					APIURL:   "api-url",
					StoreURL: "store-api-url",
					Token:    "dummy_token",
					Launcher: Launcher{},
				},
			},
			filePath: "/home/user/.sdlocal/config",
		}
	}
	t.Run("success", func(t *testing.T) {
		c := testConfig()
		c.SetCurrent("test")

		assert.Equal(t, "test", c.Current)
	})
	t.Run("failure", func(t *testing.T) {
		c := testConfig()
		err := c.SetCurrent("unknownconfig")

		assert.Equal(t, "config `unknownconfig` does not exist", err.Error())
	})
}

func TestConfigSave(t *testing.T) {
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

		config := Config{
			Current: "default",
			Entries: map[string]*Entry{
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

		err = config.Save()
		if err != nil {
			t.Fatal(err)
		}

		actual, err := ioutil.ReadFile(cnfPath)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := yaml.Marshal(config)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expected, actual)
	})
}

func TestSetEntry(t *testing.T) {
	testCases := []struct {
		name        string
		setting     map[string]string
		expectEntry Entry
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
			expectEntry: Entry{
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
			expectEntry: Entry{
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
			expectEntry: Entry{
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
			expectEntry: Entry{
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

			e := &Entry{}

			for key, val := range tt.setting {
				err := e.Set(key, val)
				if key == "invalidKey" {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}
			assert.Equal(t, tt.expectEntry, *e)
		})
	}
}
