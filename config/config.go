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

// Config is entity struct of sd-local config
type Config struct {
	APIURL   string   `yaml:"api-url"`
	StoreURL string   `yaml:"store-url"`
	Token    string   `yaml:"token"`
	Launcher Launcher `yaml:"launcher"`
}

// ConfigList is a set of sd-local config entities
type ConfigList struct {
	Configs  map[string]*Config `yaml:"configs"`
	Current  string             `yaml:"current"`
	filePath string             `yaml:"-"`
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

	err = yaml.NewEncoder(file).Encode(ConfigList{
		Configs: map[string]*Config{
			"default": newConfig(),
		},
		Current: "default",
	})
	if err != nil {
		return err
	}

	return nil
}

func newConfig() *Config {
	return &Config{
		Launcher: Launcher{
			Version: "stable",
			Image:   "screwdrivercd/launcher",
		},
	}
}

func NewConfigList(configPath string) (ConfigList, error) {
	err := create(configPath)
	if err != nil {
		return ConfigList{}, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return ConfigList{}, fmt.Errorf("failed to read config file: %v", err)
	}

	var c = ConfigList{
		filePath: configPath,
	}

	err = yaml.NewDecoder(file).Decode(&c)
	if err != nil {
		return ConfigList{}, fmt.Errorf("failed to parse config file: %v", err)
	}

	return c, nil
}

func (c *ConfigList) Add(name string) error {
	_, exist := c.Configs[name]
	if exist {
		return fmt.Errorf("config `%s` already exists", name)
	}

	c.Configs[name] = newConfig()
	return nil
}

func (c *ConfigList) Get(name string) (*Config, error) {
	currentConfig, exists := c.Configs[name]
	if !exists {
		return &Config{}, fmt.Errorf("config `%s` does not exist", name)
	}

	return currentConfig, nil
}

func (c *ConfigList) Save() error {
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
func (c *Config) Set(key, value string) error {
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
