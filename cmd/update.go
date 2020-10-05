package cmd

import (
	"bufio"
	"fmt"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
	"os"
)

func selfUpdate() {
	current := version
	latest, found, err := selfupdate.DetectLatest("screwdriver-cd/sd-local")
	if err != nil {
		fmt.Println("Error occurred while detecting version:", err)
		return
	}
	fmt.Println("Current version:", current)
	if current == "dev" {
		return
	}
	v := semver.MustParse(current)
	if !found || latest.Version.LTE(v) {
		fmt.Println("Current version is the latest")
		return
	}
	fmt.Print("Do you want to update to ", latest.Version, "? (y/n): ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		fmt.Println("Invalid input")
		return
	}
	if input == "n\n" {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("Could not locate executable path")
		return
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		fmt.Println("Error occurred while updating binary:", err)
		return
	}
	fmt.Println("Successfully updated to version", latest.Version)
}
func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update to the latest version",
		Run: func(cmd *cobra.Command, args []string) {
			selfUpdate()
		},
	}
}
