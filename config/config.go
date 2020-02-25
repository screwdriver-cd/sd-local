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
	filePath string   `yaml:"-"`
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
		Launcher: Launcher{
			Version: "stable",
			Image:   "screwdrivercd/launcher",
		},
	})
	if err != nil {
		return err
	}

	return nil
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

	var config = Config{
		filePath: configPath,
	}

	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
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

	file, err := os.OpenFile(c.filePath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	err = yaml.NewEncoder(file).Encode(c)
	if err != nil {
		return err
	}

	return nil
}
