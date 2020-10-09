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
	updateFlag   = false
	detectLatest = selfupdate.DetectLatest
	updateTo     = selfupdate.UpdateTo
)

func getLatestVersion() (*selfupdate.Release, error) {
	latest, found, err := detectLatest(githubSlug)

	if err != nil {
		return &selfupdate.Release{}, err
	}
	if !found {
		return &selfupdate.Release{}, errors.New("Repositry Not Found")
	}

	return latest, nil
}

func canUpdate(latest *selfupdate.Release) (bool, error) {
	currentVersion := version
	logrus.Info("Current version: ", currentVersion)

	if currentVersion == "dev" {
		return true, errors.New("This is a development version and cannot be updated")
	}

	v := semver.MustParse(currentVersion)

	if latest.Version.LTE(v) {
		logrus.Warn("Current version is latest")
		return true, nil
	}
	return false, nil
}

func isAborted(input string) (aborted bool, err error) {
	if input == "y" || input == "Y" || input == "yes" || input == "Yes" {
		return false, nil
	}
	if input == "n" || input == "N" || input == "no" || input == "No" || input == "" {
		logrus.Warn("Aborted the update")
		return true, nil
	}
	return true, errors.New("Invalid input")
}

func selfUpdate() error {
	latestVersion, err := getLatestVersion()
	if err != nil {
		return err
	}

	aborted, err := canUpdate(latestVersion)
	if err != nil || aborted {
		return err
	}

	if !updateFlag {
		fmt.Print("Do you want to update to ", latestVersion.Version.String(), "? [y/N]: ")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSuffix(input, "\n")
		aborted, err := isAborted(input)
		if aborted {
			return err
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	logrus.Info("Updating ...")
	if err := updateTo(latestVersion.AssetURL, exe); err != nil {
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
