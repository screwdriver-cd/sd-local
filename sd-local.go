package main

import (
	"log"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/screwdriver-cd/sd-local/screwdriver"
)

const (
	jobName string = "main"
)

func main() {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	configPath := path.Join(homeDir, "/.sdlocal/config")

	config, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	api, err := screwdriver.New(config.APIURL, config.Token)
	if err != nil {
		log.Fatal(err)
	}

	job, err := api.Job(jobName, "./screwdriver.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = launch.New(job, config, jobName, api.JWT()).Run()
	if err != nil {
		log.Fatal(err)
	}
}
