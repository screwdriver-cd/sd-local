package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const githubSlug = "screwdriver-cd/sd-local"

var (
	currentVersion = "1.0.4"
	updateFlag     = false
)

func canUpdate() (*selfupdate.Release, error) {

	if currentVersion == "dev" {
		return &selfupdate.Release{}, errors.New("This is a development version and cannot be updated")
	}

	latest, found, err := selfupdate.DetectLatest(githubSlug)
	if err != nil {
		return &selfupdate.Release{}, err
	}
	if !found {
		return &selfupdate.Release{}, errors.New("Repositry Not Found")
	}
	v := semver.MustParse(currentVersion)

	if latest.Version.LTE(v) {
		return &selfupdate.Release{}, errors.New("Current version is latest")
	}

	return latest, nil
}

func checkUserInput(input string) error {
	if input == "y" || input == "Y" {
		return nil
	}

	if input == "n" || input == "N" || input == "" {
		return errors.New("Aborted")
	}
	return errors.New("Invalid input")
}

func selfUpdate() error {
	latestVersion, err := canUpdate()
	if err != nil {
		return err
	}

	logrus.Info("Current version:", currentVersion)
	if !updateFlag {
		fmt.Print("Do you want to update to", latestVersion.Version.String(), "? [y/N]: ")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSuffix(input, "\n")
		err = checkUserInput(input)
		if err != nil {
			return err
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	logrus.Info("Updating ...")
	if err := selfupdate.UpdateTo(latestVersion.AssetURL, exe); err != nil {
		return err
	}
	logrus.Info("Successfully updated to version ", latestVersion.Version)
	return nil
}

func newUpdateCmd() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return selfUpdate()
		},
	}
	updateCmd.Flags().BoolVarP(&updateFlag, "yes", "y", false, "answer yes for all questions")

	return updateCmd
}
