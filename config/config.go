package config

import (
	"fmt"
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

// Launcher is launcher entity struct
type Launcher struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
}

// Config is successConfig entity struct
type Config struct {
	APIURL   string   `yaml:"api-url"`
	StoreURL string   `yaml:"store-url"`
	Token    string   `yaml:"token"`
	Launcher Launcher `yaml:"launcher"`
}

// ReadConfig returns parsed successConfig file
func ReadConfig(configPath string) (Config, error) {
	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read successConfig file: %v", err)
	}

	config := Config{}

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse successConfig file: %v\ncontents:\n\n%v", err, string(buf))
	}

	return config, nil
}
