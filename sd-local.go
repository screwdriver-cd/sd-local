package main

import (
	"log"

	"github.com/screwdriver-cd/sd-local/cmd"
	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/screwdriver-cd/sd-local/screwdriver"
)

func main() {
	path := "file/path"
	config, err := config.ReadConfig(path)
	if err != nil {
		log.Fatal(err)
	}

	api, err := screwdriver.New(config.APIURL, config.Token)
	if err != nil {
		log.Fatal(err)
	}

	job, err := api.Job("main", "./screwdriver.yaml")
	if err != nil {
		log.Fatal(err)
	}
	// lauch.Newには API interfaceを渡すように修正したい。
	l := launch.New(job, config, "main", api.JWT())

	if err := cmd.Execute(config, api, l); err != nil {
		log.Fatal(err)
	}
}
