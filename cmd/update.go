package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	githubSlug     = "screwdriver-cd/sd-local"
	currentVersion = version
	updateFlag     = false
)

func canUpdate() (*selfupdate.Release, error) {

	if currentVersion == "dev" {
		err := errors.New("This is a development command and cannot be updated.")
		return &selfupdate.Release{}, err
	}

	latest, found, err := selfupdate.DetectLatest(githubSlug)
	if err != nil {
		return &selfupdate.Release{}, err
	}
	if !found {
		err = errors.New("Repositry Not Found")
		return &selfupdate.Release{}, err
	}
	v := semver.MustParse(currentVersion)
	if latest.Version.LTE(v) {
		err = errors.New("Current version is latest")
		return &selfupdate.Release{}, err
	}
	return latest, err
}

func askUpdateForUser(latestVersion *selfupdate.Release) bool {
	fmt.Print("Do you want to update to ", latestVersion.Version.String(), "? (y/n): ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		logrus.Error("Invalid input")
		return false
	}
	if input == "n\n" {
		logrus.Warn("Aborted")
		return false
	}
	return true
}

func selfUpdate() error {
	latestVersion, err := canUpdate()
	if err != nil {
		return err
	}

	logrus.Info("Current version:", currentVersion)
	if !updateFlag {
		ok := askUpdateForUser(latestVersion)
		if !ok {
			return nil
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
			err := selfUpdate()
			if err != nil {
				return err
			}
			return nil
		},
	}
	updateCmd.Flags().BoolVarP(&updateFlag, "yes", "y", false, "answer yes for all questions")

	return updateCmd
}
