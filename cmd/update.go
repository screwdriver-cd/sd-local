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

var (
	githubSlug     = "screwdriver-cd/sd-local"
	currentVersion = version
	updateFlag     = false
)

func canUpdate() (*selfupdate.Release, error) {

	if currentVersion == "dev" {
		return &selfupdate.Release{}, errors.New("This is a development command and cannot be updated.")
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

func askUpdateForUser(latestVersion *selfupdate.Release) error {
	fmt.Print("Do you want to update to", latestVersion.Version.String(), "? [y/N]: ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return err
	}
	input = strings.TrimSuffix(input, "\n")
	if input == "y" {
		return nil
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
		err := askUpdateForUser(latestVersion)
		if err != nil {
			return err
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	logrus.Info("Updateing ...")
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
