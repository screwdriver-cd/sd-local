package launch

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeExecCommand struct {
	id      string
	execCmd func(command string, args ...string) *exec.Cmd
}

func newFakeExecCommand(id string) *fakeExecCommand {
	c := &fakeExecCommand{
		id: id,
		execCmd: func(name string, args ...string) *exec.Cmd {
			cs := []string{"-test.run=TestHelperProcess", "--", name}
			cs = append(cs, args...)
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_TEST_MODE=%s", id)}
			return cmd
		},
	}
	return c
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
	defer func() {
		execCommand = exec.Command
	}()

	d := &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        "launcher",
		setupImageVersion: "latest",
	}

	testCase := []struct {
		name        string
		id          string
		expectError error
	}{
		{"success", "SUCCESS_SETUP_BIN", nil},
		{"failure volume create", "FAIL_CREATING_VOLUME", fmt.Errorf("failed to create docker volume")},
		{"failure container run", "FAIL_CONTAINER_RUN", fmt.Errorf("failed to prepare build scripts")},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeExecCommand(tt.id)
			execCommand = c.execCmd
			err := d.setupBin()

			assert.Equal(t, tt.expectError, err)
		})
	}
}

func TestRunBuild(t *testing.T) {
	buildConfig := newBuildConfig()

	d := &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        "launcher",
		setupImageVersion: "latest",
	}

	testCase := []struct {
		name        string
		id          string
		expectError error
	}{
		{"success", "SUCCESS_RUN_BUILD", nil},
		{"fail run build", "FAIL_BUILD_CONTAINER_RUN", fmt.Errorf("exit status 1")},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeExecCommand(tt.id)
			execCommand = c.execCmd
			err := d.runBuild(buildConfig)
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
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

	cmd, subcmd, args := args[0], args[1], args[2:]
	_, _ = cmd, args

	switch os.Getenv("GO_TEST_MODE") {
	case "":
		os.Exit(1)
	case "SUCCESS_SETUP_BIN":
		os.Exit(0)
	case "FAIL_CREATING_VOLUME":
		os.Exit(1)
	case "FAIL_CONTAINER_RUN":
		if subcmd == "volume" {
			os.Exit(0)
		}
		os.Exit(1)
	case "SUCCESS_RUN_BUILD":
		os.Exit(0)
	case "FAIL_BUILD_CONTAINER_RUN":
		os.Exit(1)
	}
}
