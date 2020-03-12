package launch

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeExecCommand struct {
	id       string
	execCmd  func(command string, args ...string) *exec.Cmd
	commands []string
}

func newFakeExecCommand(id string) *fakeExecCommand {
	c := &fakeExecCommand{}
	c.id = id
	c.commands = make([]string, 0, 5)
	c.execCmd = func(name string, args ...string) *exec.Cmd {
		c.commands = append(c.commands, fmt.Sprintf("%s %s", name, strings.Join(args, " ")))
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_TEST_MODE=%s", id)}
		return cmd
	}
	return c
}

func TestNewDocker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &docker{
			volume:            "SD_LAUNCH_BIN",
			setupImage:        "launcher",
			setupImageVersion: "latest",
			useSudo:           false,
		}

		d := newDocker("launcher", "latest", false)

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
		{"failure volume create", "FAIL_CREATING_VOLUME", fmt.Errorf("failed to create docker volume: exit status 1")},
		{"failure container run", "FAIL_CONTAINER_RUN", fmt.Errorf("failed to prepare build scripts: exit status 1")},
		{"failure launcher image pull", "FAIL_LAUNCHER_PULL", fmt.Errorf("failed to pull launcher image: exit status 1")},
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

func TestSetupBinWithSudo(t *testing.T) {
	defer func() {
		execCommand = exec.Command
	}()

	d := &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        "launcher",
		setupImageVersion: "latest",
		useSudo:           true,
	}

	testCase := []struct {
		name        string
		id          string
		expectError error
	}{
		{"success", "SUCCESS_SETUP_BIN_SUDO", nil},
		{"failure volume create", "FAIL_CREATING_VOLUME_SUDO", fmt.Errorf("failed to create docker volume: exit status 1")},
		{"failure container run", "FAIL_CONTAINER_RUN_SUDO", fmt.Errorf("failed to prepare build scripts: exit status 1")},
		{"failure launcher image pull", "FAIL_LAUNCHER_PULL_SUDO", fmt.Errorf("failed to pull launcher image: exit status 1")},
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
	defer func() {
		execCommand = exec.Command
	}()

	d := &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        "launcher",
		setupImageVersion: "latest",
	}

	testCase := []struct {
		name             string
		id               string
		expectError      error
		expectedCommands []string
		buildConfig      buildConfig
	}{
		{"success", "SUCCESS_RUN_BUILD", nil,
			[]string{
				"docker pull node:12",
				fmt.Sprintf("docker container run --rm -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd node:12 /opt/sd/local_run.sh ", d.volume)},
			newBuildConfig()},
		{"success with memory limit", "SUCCESS_RUN_BUILD", nil,
			[]string{
				"docker pull node:12",
				fmt.Sprintf("docker container run -m2GB --rm -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd node:12 /opt/sd/local_run.sh ", d.volume)},
			newBuildConfig(func(b *buildConfig) {
				b.MemoryLimit = "2GB"
			})},
		{"failure build run", "FAIL_BUILD_CONTAINER_RUN", fmt.Errorf("failed to run build container: exit status 1"), []string{}, newBuildConfig()},
		{"failure build image pull", "FAIL_BUILD_IMAGE_PULL", fmt.Errorf("failed to pull user image exit status 1"), []string{}, newBuildConfig()},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeExecCommand(tt.id)
			execCommand = c.execCmd
			err := d.runBuild(tt.buildConfig)
			for i, expectedCommand := range tt.expectedCommands {
				assert.True(t, strings.Contains(c.commands[i], expectedCommand), "expect %q \nbut got \n%q", expectedCommand, c.commands[i])
			}
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRunBuildWithSudo(t *testing.T) {
	defer func() {
		execCommand = exec.Command
	}()

	d := &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        "launcher",
		setupImageVersion: "latest",
		useSudo:           true,
	}

	testCase := []struct {
		name             string
		id               string
		expectError      error
		expectedCommands []string
		buildConfig      buildConfig
	}{
		{"success", "SUCCESS_RUN_BUILD_SUDO", nil,
			[]string{
				"docker pull node:12",
				fmt.Sprintf("sudo docker container run --rm -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd node:12 /opt/sd/local_run.sh ", d.volume)},
			newBuildConfig()},
		{"success with memory limit", "SUCCESS_RUN_BUILD_SUDO", nil,
			[]string{
				"docker pull node:12",
				fmt.Sprintf("sudo docker container run -m2GB --rm -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd node:12 /opt/sd/local_run.sh ", d.volume)},
			newBuildConfig(func(b *buildConfig) {
				b.MemoryLimit = "2GB"
			})},
		{"failure build run", "FAIL_BUILD_CONTAINER_RUN_SUDO", fmt.Errorf("failed to run build container: exit status 1"), []string{}, newBuildConfig()},
		{"failure build image pull", "FAIL_BUILD_IMAGE_PULL_SUDO", fmt.Errorf("failed to pull user image exit status 1"), []string{}, newBuildConfig()},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeExecCommand(tt.id)
			execCommand = c.execCmd
			err := d.runBuild(tt.buildConfig)
			for i, expectedCommand := range tt.expectedCommands {
				assert.True(t, strings.Contains(c.commands[i], expectedCommand), "expect %q \nbut got \n%q", expectedCommand, c.commands[i])
			}
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
	testCase := os.Getenv("GO_TEST_MODE")
	if strings.Index(testCase, "SUDO") > 0 {
		subcmd = args[0]
	}

	switch testCase {
	case "":
		os.Exit(1)
	case "SUCCESS_SETUP_BIN":
		os.Exit(0)
	case "SUCCESS_SETUP_BIN_SUDO":
		os.Exit(0)
	case "FAIL_CREATING_VOLUME":
		os.Exit(1)
	case "FAIL_CREATING_VOLUME_SUDO":
		os.Exit(1)
	case "FAIL_CONTAINER_RUN":
		if subcmd == "volume" {
			os.Exit(0)
		}
		if subcmd == "pull" {
			os.Exit(0)
		}
		os.Exit(1)
	case "FAIL_CONTAINER_RUN_SUDO":
		if subcmd == "volume" {
			os.Exit(0)
		}
		if subcmd == "pull" {
			os.Exit(0)
		}
		os.Exit(1)
	case "SUCCESS_RUN_BUILD":
		os.Exit(0)
	case "SUCCESS_RUN_BUILD_SUDO":
		os.Exit(0)
	case "FAIL_BUILD_CONTAINER_RUN":
		if subcmd == "pull" {
			os.Exit(0)
		}
		os.Exit(1)
	case "FAIL_BUILD_CONTAINER_RUN_SUDO":
		if subcmd == "pull" {
			os.Exit(0)
		}
		os.Exit(1)
	case "FAIL_LAUNCHER_PULL":
		if subcmd == "pull" {
			os.Exit(1)
		}
		os.Exit(0)
	case "FAIL_LAUNCHER_PULL_SUDO":
		if subcmd == "pull" {
			os.Exit(1)
		}
		os.Exit(0)
	case "FAIL_BUILD_IMAGE_PULL":
		if subcmd == "pull" {
			os.Exit(1)
		}
		os.Exit(0)
	case "FAIL_BUILD_IMAGE_PULL_SUDO":
		if subcmd == "pull" {
			os.Exit(1)
		}
		os.Exit(0)
	}
}
