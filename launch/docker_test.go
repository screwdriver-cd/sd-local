package launch

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

const launchImage = "launcher:latest"

func fakeExecCommand(name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestNewDocker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &docker{
			volume:            "SD_LAUNCH_BIN",
			setupImage:        "launcher",
			setupImageVersion: "latest",
		}

		d := newDocker("launcher", "latest")

		assert.Equal(t, expected, d)
	})
}

func TestSetupBin(t *testing.T) {
	execCommand = fakeExecCommand

	defer func() {
		execCommand = exec.Command
	}()

	d := &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        "launcher",
		setupImageVersion: "latest",
	}

	t.Run("success", func(t *testing.T) {
		err := d.SetupBin()

		assert.Equal(t, nil, err)
	})

	t.Run("failure volume create", func(t *testing.T) {
		d.volume = "TO_FAIL_CREATING_VOLUME"
		defer func() {
			d.volume = "SD_LAUNCH_BIN"
		}()

		err := d.SetupBin()

		assert.Equal(t, fmt.Errorf("failed to create docker volume"), err)
	})

	t.Run("failure container run", func(t *testing.T) {
		d.setupImage = "TO_FAIL_CONTAINER_RUN"
		defer func() {
			d.setupImage = "launcher"
		}()

		err := d.SetupBin()

		assert.Equal(t, fmt.Errorf("failed to prepare build scripts"), err)
	})

}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "no command\n")
		os.Exit(2)
	}

	cmd, subcmd, subsubcmd, args := args[0], args[1], args[2], args[3:]

	if cmd != "sudo" {
		fmt.Fprintf(os.Stderr, "expected 'sudo', but %v\n", cmd)
		os.Exit(1)
	}

	if subcmd != "docker" {
		fmt.Fprintf(os.Stderr, "expected 'docker', but %v\n", subcmd)
		os.Exit(1)
	}

	switch subsubcmd {
	case "volume":
		expectedArgs := []string{"create", "--name", "SD_LAUNCH_BIN"}

		if !assert.Equal(t, expectedArgs, args) {
			fmt.Fprintf(os.Stderr, "expected args: %v but actual: %v", expectedArgs, args)
			os.Exit(1)
		}
	case "container":
		expectedArgs := []string{"run", "-v", "SD_LAUNCH_BIN:/opt/sd/", launchImage, "--entrypoint", "/bin/echo set up bin"}
		if !assert.Equal(t, expectedArgs, args) {
			fmt.Fprintf(os.Stderr, "expected args: %v but actual: %v", expectedArgs, args)
			os.Exit(1)
		}
	}
}
