package main

import (
	"fmt"
	"github.com/screwdriver-cd/sd-local/config"
	"log"
	"os"
	"path"
)

func main() {
	path := path.Join(os.Getenv("HOME"), ".sdlocal")
	config, err := config.ReadConfig(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(config)
}
