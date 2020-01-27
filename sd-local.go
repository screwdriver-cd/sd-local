package main

import (
	"log"

	"github.com/screwdriver-cd/sd-local/cmd"
)

func main() {
	path := "file/path"
	config, err := config.ReadConfig(path)
	if err != nil {
		log.Fatal(err)
	}

	api, err := screwdriver.New(config.ApiUrl, config.Token)
	if err != nil {
		log.Fatal(err)
	}

	l, err := launch.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Execute(config, api, l); err != nil {
		log.Fatal(err)
	}
}
