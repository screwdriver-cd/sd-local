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

func dummyEntry() *Entry {
	return &Entry{
		APIURL:   "api-url",
		StoreURL: "store-api-url",
		Token:    "dummy_token",
		Launcher: Launcher{
			Version: "latest",
			Image:   "screwdrivercd/launcher",
		},
	}
}

// dummyConfig is eqeual to ./testdata/successConfig
func dummyConfig() Config {
	return Config{
		Entries: map[string]*Entry{
			"default": dummyEntry(),
		},
		Current: "default",
	}
}

func TestCreateConfig(t *testing.T) {
	t.Run("success to init config", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		cnfPath := filepath.Join(testDir, fmt.Sprintf("%vconfig", rand.Int()))
		defer os.Remove(cnfPath)

		expect := Config{
			Entries: map[string]*Entry{
				"default": DefaultEntry(),
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
				"default": DefaultEntry(),
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

		testConfig := dummyConfig()
		testConfig.filePath = cnfPath

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
	cases := map[string]struct {
		current     string
		expectEntry *Entry
		expectErr   error
	}{
		"success": {
			current:     "default",
			expectEntry: dummyEntry(),
			expectErr:   nil,
		},
		"failed": {
			current:     "doesnotexist",
			expectEntry: &Entry{},
			expectErr:   fmt.Errorf("config `doesnotexist` does not exist"),
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			config := Config{
				Entries: map[string]*Entry{
					"default": dummyEntry(),
				},
				Current: "default",
			}
			actual, err := config.Entry(test.current)

			assert.Equal(t, test.expectErr, err)
			assert.Equal(t, test.expectEntry, actual)
		})
	}
}

func TestConfigAddEntry(t *testing.T) {
	cases := map[string]struct {
		addedEntryName string
		expectConfig   Config
		expectErr      error
	}{
		"successfully added a test entry": {
			addedEntryName: "test",
			expectConfig: Config{
				Entries: map[string]*Entry{
					"default": dummyEntry(),
					"test":    DefaultEntry(),
				},
				Current: "default",
			},
			expectErr: nil,
		},
		"failure by the name that exists": {
			addedEntryName: "default",
			expectConfig: Config{
				Entries: map[string]*Entry{
					"default": dummyEntry(),
				},
				Current: "default",
			},
			expectErr: fmt.Errorf("config `default` already exists"),
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			config := dummyConfig()
			err := config.AddEntry(test.addedEntryName, DefaultEntry())
			assert.Equal(t, test.expectErr, err)
			assert.Equal(t, test.expectConfig, config)
		})
	}
}

func TestConfigDeleteEntry(t *testing.T) {
	cases := map[string]struct {
		deletedEntryName string
		expectConfig     Config
		expectErr        error
	}{
		"successfully deleted a test entry": {
			deletedEntryName: "test",
			expectConfig: Config{
				Entries: map[string]*Entry{
					"default": dummyEntry(),
				},
				Current: "default",
			},
			expectErr: nil,
		},
		"failure by the name that does not exist": {
			deletedEntryName: "doesnotexist",
			expectConfig: Config{
				Entries: map[string]*Entry{
					"default": dummyEntry(),
					"test":    DefaultEntry(),
				},
				Current: "default",
			},
			expectErr: fmt.Errorf("config `doesnotexist` does not exist"),
		},
		"failure by trying to delete current entry": {
			deletedEntryName: "default",
			expectConfig: Config{
				Entries: map[string]*Entry{
					"default": dummyEntry(),
					"test":    DefaultEntry(),
				},
				Current: "default",
			},
			expectErr: fmt.Errorf("config `default` is current config"),
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := Config{
				Entries: map[string]*Entry{
					"default": dummyEntry(),
					"test":    DefaultEntry(),
				},
				Current: "default",
			}

			err := config.DeleteEntry(test.deletedEntryName)
			assert.Equal(t, test.expectErr, err)
			assert.Equal(t, test.expectConfig, config)
		})
	}
}

func TestConfigSetCurrent(t *testing.T) {
	cases := map[string]struct {
		setEntryName  string
		expectCurrent string
		expectErr     error
	}{
		"success to set": {
			setEntryName:  "test",
			expectCurrent: "test",
			expectErr:     nil,
		},
		"failure to set": {
			setEntryName:  "doesnotexist",
			expectCurrent: "default",
			expectErr:     fmt.Errorf("config `doesnotexist` does not exist"),
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			config := Config{
				Current: "default",
				Entries: map[string]*Entry{
					"default": dummyEntry(),
					"test":    DefaultEntry(),
				},
			}
			err := config.SetCurrent(test.setEntryName)
			assert.Equal(t, test.expectErr, err)
			assert.Equal(t, test.expectCurrent, config.Current)
		})
	}
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
