package main

import (
	"log"

	"github.com/screwdriver-cd/sd-local/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
