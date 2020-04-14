package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-yaml/yaml"
)

// Launcher is launcher entity struct
type Launcher struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
}

// Entry is entity struct of sd-local config
type Entry struct {
	APIURL   string   `yaml:"api-url"`
	StoreURL string   `yaml:"store-url"`
	Token    string   `yaml:"token"`
	Launcher Launcher `yaml:"launcher"`
}

// Config is a set of sd-local config entities
type Config struct {
	Entries  map[string]*Entry `yaml:"configs"`
	Current  string            `yaml:"current"`
	filePath string            `yaml:"-"`
}

func create(configPath string) error {
	_, err := os.Stat(configPath)
	// if file exists return nil
	if err == nil {
		return nil
	}

	dir := filepath.Dir(configPath)
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = yaml.NewEncoder(file).Encode(Config{
		Entries: map[string]*Entry{
			"default": newEntry(),
		},
		Current: "default",
	})
	if err != nil {
		return err
	}

	return nil
}

func newEntry() *Entry {
	return &Entry{
		Launcher: Launcher{
			Version: "stable",
			Image:   "screwdrivercd/launcher",
		},
	}
}

func New(configPath string) (Config, error) {
	err := create(configPath)
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %v", err)
	}

	var c = Config{
		filePath: configPath,
	}

	err = yaml.NewDecoder(file).Decode(&c)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %v", err)
	}

	return c, nil
}

func (c *Config) Add(name string) error {
	_, exist := c.Entries[name]
	if exist {
		return fmt.Errorf("config `%s` already exists", name)
	}

	c.Entries[name] = newEntry()
	return nil
}

func (c *Config) Get(name string) (*Entry, error) {
	currentEntry, exists := c.Entries[name]
	if !exists {
		return &Entry{}, fmt.Errorf("config `%s` does not exist", name)
	}

	return currentEntry, nil
}

func (c *Config) Save() error {
	file, err := os.OpenFile(c.filePath, os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	err = yaml.NewEncoder(file).Encode(c)
	if err != nil {
		return err
	}

	return nil
}

// Set preserve sd-local config with new value.
func (c *Entry) Set(key, value string) error {
	switch key {
	case "api-url":
		c.APIURL = value
	case "store-url":
		c.StoreURL = value
	case "token":
		c.Token = value
	case "launcher-version":
		if value == "" {
			value = "stable"
		}
		c.Launcher.Version = value
	case "launcher-image":
		if value == "" {
			value = "screwdrivercd/launcher"
		}
		c.Launcher.Image = value
	default:
		return fmt.Errorf("invalid key %s", key)
	}

	return nil
}
