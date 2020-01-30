package main

import (
	"fmt"
	"log"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/screwdriver-cd/sd-local/config"
)

const (
	configDir string = ".sdlocal"
	configName string = "config"
)

func main() {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	path := path.Join(homeDir, configDir, configName)
	config, err := config.ReadConfig(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(config)
}
