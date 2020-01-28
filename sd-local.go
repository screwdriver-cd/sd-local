package main

import (
	"fmt"

	"github.com/screwdriver-cd/sd-local/launch"
	"github.com/screwdriver-cd/sd-local/screwdriver"
)

func main() {
	job := screwdriver.Job{
		Environment: map[string]string{"FOO": "BAR"},
		Image:       "",
		Steps:       []screwdriver.Step{},
	}
	launch.New(job, "hogehoge").Run()
	fmt.Println("hello world")
}
