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

// New returns parsed config
func New(configPath string) (Config, error) {
	err := create(configPath)
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %v", err)
	}
	defer file.Close()

	var c = Config{
		filePath: configPath,
	}

	err = yaml.NewDecoder(file).Decode(&c)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %v", err)
	}

	if c.Entries == nil {
		c.Entries = make(map[string]*Entry, 0)
	}

	return c, nil
}

// AddEntry create new Entry and add it to Config
func (c *Config) AddEntry(name string) error {
	_, exist := c.Entries[name]
	if exist {
		return fmt.Errorf("config `%s` already exists", name)
	}

	c.Entries[name] = newEntry()
	return nil
}

// Entry returns an Entry object named `name`
func (c *Config) Entry(name string) (*Entry, error) {
	entry, exists := c.Entries[name]
	if !exists {
		return &Entry{}, fmt.Errorf("config `%s` does not exist", name)
	}

	return entry, nil
}

// DeleteEntry deletes Entry object named `name`
func (c *Config) DeleteEntry(name string) error {
	if name == c.Current {
		return fmt.Errorf("config `%s` is current config", name)
	}
	_, exist := c.Entries[name]
	if !exist {
		return fmt.Errorf("config `%s` does not exist", name)
	}
	delete(c.Entries, name)
	return nil
}

// SetCurrent set a specified entry as current config
func (c *Config) SetCurrent(name string) error {
	_, err := c.Entry(name)
	if err != nil {
		return err
	}

	c.Current = name

	return nil
}

// Save write Config to config file
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
func (e *Entry) Set(key, value string) error {
	switch key {
	case "api-url":
		e.APIURL = value
	case "store-url":
		e.StoreURL = value
	case "token":
		e.Token = value
	case "launcher-version":
		if value == "" {
			value = "stable"
		}
		e.Launcher.Version = value
	case "launcher-image":
		if value == "" {
			value = "screwdrivercd/launcher"
		}
		e.Launcher.Image = value
	default:
		return fmt.Errorf("invalid key %s", key)
	}

	return nil
}
