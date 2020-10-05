package main

import (
	"os"

	"github.com/screwdriver-cd/sd-local/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	textFormatter := new(logrus.TextFormatter)
	textFormatter.PadLevelText = true
	logrus.SetFormatter(textFormatter)

	if err := cmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
