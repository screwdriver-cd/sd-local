package launch

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var sshSocket = os.Getenv("SSH_AUTH_SOCK") + ":/tmp/auth.sock:rw"

const (
	fakeProcessLifeTime = 100 * time.Second
	waitForKillTime     = 100 * time.Millisecond
)

type fakeExecCommand struct {
	id       string
	execCmd  func(command string, args ...string) *exec.Cmd
	commands []string
}

type fakeOs struct {
	dirPaths  string
	fileNames string
	mkdirAll  func(path string, perm fs.FileMode) error
	WriteFile func(path string, data []byte, perm fs.FileMode) error
}

type mockInteract struct {
	Interacter
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

func newFakeOsMkdir(t *testing.T) *fakeOs {
	f := &fakeOs{}
	f.mkdirAll = func(path string, perm fs.FileMode) error {
		f.dirPaths += path + ","

		assert.Equal(t, os.FileMode(0777), perm)

		return nil
	}
	f.WriteFile = func(path string, data []byte, perm fs.FileMode) error {
		f.fileNames += path + ","

		assert.Equal(t, os.FileMode(0755), perm)

		return nil
	}

	return f
}

func (d *mockInteract) Run(c *exec.Cmd, commands [][]string) error {
	return c.Run()
}

func TestNewDocker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &docker{
			volume:            "SD_LAUNCH_BIN",
			habVolume:         "SD_LAUNCH_HAB",
			setupImage:        "launcher",
			setupImageVersion: "latest",
			useSudo:           false,
			interactiveMode:   false,
			sdUtilsPath:       ".sd-utils",
			commands:          make([]*exec.Cmd, 0, 10),
			mutex:             &sync.Mutex{},
			flagVerbose:       false,
			interact:          &Interact{},
			socketPath:        "/auth.sock",
			localVolumes:      []string{"path:path"},
			buildUser:         "jithin",
			noImagePull:       false,
			dind: DinD{
				enabled:         true,
				volume:          "SD_DIND_CERT",
				shareVolumeName: "SD_DIND_SHARE",
				shareVolumePath: "/opt/sd_dind_share",
				container:       "sd-local-dind",
				network:         "sd-local-dind-bridge",
				image:           "docker:23.0.1-dind-rootless",
			},
		}

		d := newDocker("launcher", "latest", false, false, ".sd-utils", "/auth.sock", false, []string{"path:path"}, "jithin", false, true)

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
		socketPath:        os.Getenv("SSH_AUTH_SOCK"),
		dind: DinD{
			enabled:         false,
			volume:          "SD_DIND_CERT",
			shareVolumeName: "SD_DIND_SHARE",
			shareVolumePath: "/opt/sd_dind_share",
			container:       "sd-local-dind",
			network:         "sd-local-dind-bridge",
			image:           "docker:23.0.1-dind-rootless",
		},
	}

	testCase := []struct {
		name             string
		id               string
		expectError      error
		expectedCommands []string
		buildEntry       buildEntry
	}{
		{"success", "SUCCESS_RUN_BUILD", nil,
			[]string{
				"docker pull node:12",
				fmt.Sprintf("docker container run --rm --entrypoint /bin/sh -e SSH_AUTH_SOCK=/tmp/auth.sock -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd -v %s:/opt/sd/hab -v %s --pull never node:12 /opt/sd/local_run.sh ", d.volume, d.habVolume, sshSocket)},
			newBuildEntry()},
		{"success with memory limit", "SUCCESS_RUN_BUILD", nil,
			[]string{
				"docker pull node:12",
				fmt.Sprintf("docker container run --rm --entrypoint /bin/sh -e SSH_AUTH_SOCK=/tmp/auth.sock -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd -v %s:/opt/sd/hab -v %s -m2GB --pull never node:12 /opt/sd/local_run.sh ", d.volume, d.habVolume, sshSocket)},
			newBuildEntry(func(b *buildEntry) {
				b.MemoryLimit = "2GB"
			})},
		{"failure build run", "FAIL_BUILD_CONTAINER_RUN", fmt.Errorf("failed to run build container: exit status 1"), []string{}, newBuildEntry()},
		{"failure build image pull", "FAIL_BUILD_IMAGE_PULL", fmt.Errorf("failed to pull user image exit status 1"), []string{}, newBuildEntry()},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeExecCommand(tt.id)
			execCommand = c.execCmd
			err := d.runBuild(tt.buildEntry)
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

func TestRunBuildWithDind(t *testing.T) {
	defer func() {
		execCommand = exec.Command
	}()

	d := &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        "launcher",
		setupImageVersion: "latest",
		socketPath:        os.Getenv("SSH_AUTH_SOCK"),
		dind: DinD{
			enabled:         true,
			volume:          "SD_DIND_CERT",
			shareVolumeName: "SD_DIND_SHARE",
			shareVolumePath: "/opt/sd_dind_share",
			container:       "sd-local-dind",
			network:         "sd-local-dind-bridge",
			image:           "docker:23.0.1-dind-rootless",
		},
	}

	testCase := []struct {
		name             string
		id               string
		expectError      error
		expectedCommands []string
		buildEntry       buildEntry
	}{
		{"success with dind", "SUCCESS_RUN_BUILD", nil,
			[]string{
				"docker pull docker:23.0.1-dind-rootless",
				"docker network create sd-local-dind-bridge",
				"docker container run --rm --privileged --pull never --name sd-local-dind -d --network sd-local-dind-bridge --network-alias docker -e DOCKER_TLS_CERTDIR=/certs -v SD_DIND_CERT:/certs/client -v SD_DIND_SHARE:/opt/sd_dind_share docker:23.0.1-dind-rootless",
				"docker pull node:12",
				fmt.Sprintf("docker container run --network %s -e DOCKER_TLS_CERTDIR=/certs -e DOCKER_HOST=tcp://docker:2376 -e DOCKER_TLS_VERIFY=1 -e DOCKER_CERT_PATH=/certs/client -e SD_DIND_SHARE_PATH=%s -v %s:/certs/client:ro -v %s:%s --rm --entrypoint /bin/sh -e SSH_AUTH_SOCK=/tmp/auth.sock -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd -v %s:/opt/sd/hab -v %s --pull never node:12 /opt/sd/local_run.sh ", d.dind.network, d.dind.shareVolumePath, d.dind.volume, d.dind.shareVolumeName, d.dind.shareVolumePath, d.volume, d.habVolume, sshSocket)},
			newBuildEntry(func(b *buildEntry) {
				b.Annotations["screwdriver.cd/dockerEnabled"] = true
			})},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeExecCommand(tt.id)
			execCommand = c.execCmd
			err := d.runBuild(tt.buildEntry)
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

func TestRunBuildWithNoImagePull(t *testing.T) {
	defer func() {
		execCommand = exec.Command
	}()

	d := &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        "launcher",
		setupImageVersion: "latest",
		socketPath:        os.Getenv("SSH_AUTH_SOCK"),
		noImagePull:       true,
		dind: DinD{
			enabled:         true,
			volume:          "SD_DIND_CERT",
			shareVolumeName: "SD_DIND_SHARE",
			shareVolumePath: "/opt/sd_dind_share",
			container:       "sd-local-dind",
			network:         "sd-local-dind-bridge",
			image:           "docker:23.0.1-dind-rootless",
		},
	}

	testCase := []struct {
		name             string
		id               string
		expectError      error
		expectedCommands []string
		buildEntry       buildEntry
	}{
		{"success with no image pull", "SUCCESS_RUN_BUILD", nil,
			[]string{
				"docker network create sd-local-dind-bridge",
				"docker container run --rm --privileged --pull never --name sd-local-dind -d --network sd-local-dind-bridge --network-alias docker -e DOCKER_TLS_CERTDIR=/certs -v SD_DIND_CERT:/certs/client -v SD_DIND_SHARE:/opt/sd_dind_share docker:23.0.1-dind-rootless",
				fmt.Sprintf("docker container run --network %s -e DOCKER_TLS_CERTDIR=/certs -e DOCKER_HOST=tcp://docker:2376 -e DOCKER_TLS_VERIFY=1 -e DOCKER_CERT_PATH=/certs/client -e SD_DIND_SHARE_PATH=%s -v %s:/certs/client:ro -v %s:%s --rm --entrypoint /bin/sh -e SSH_AUTH_SOCK=/tmp/auth.sock -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd -v %s:/opt/sd/hab -v %s --pull never node:12 /opt/sd/local_run.sh ", d.dind.network, d.dind.shareVolumePath, d.dind.volume, d.dind.shareVolumeName, d.dind.shareVolumePath, d.volume, d.habVolume, sshSocket)},
			newBuildEntry(func(b *buildEntry) {
				b.Annotations["screwdriver.cd/dockerEnabled"] = true
			})},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeExecCommand(tt.id)
			execCommand = c.execCmd
			err := d.runBuild(tt.buildEntry)
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
		socketPath:        os.Getenv("SSH_AUTH_SOCK"),
		dind: DinD{
			enabled:         false,
			volume:          "SD_DIND_CERT",
			shareVolumeName: "SD_DIND_SHARE",
			shareVolumePath: "/opt/sd_dind_share",
			container:       "sd-local-dind",
			network:         "sd-local-dind-bridge",
			image:           "docker:23.0.1-dind-rootless",
		},
	}

	testCase := []struct {
		name             string
		id               string
		expectError      error
		expectedCommands []string
		buildEntry       buildEntry
	}{
		{"success", "SUCCESS_RUN_BUILD_SUDO", nil,
			[]string{
				"sudo docker pull node:12",
				fmt.Sprintf("sudo docker container run --rm --entrypoint /bin/sh -e SSH_AUTH_SOCK=/tmp/auth.sock -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd -v %s:/opt/sd/hab -v %s --pull never node:12 /opt/sd/local_run.sh ", d.volume, d.habVolume, sshSocket)},
			newBuildEntry()},
		{"success with memory limit", "SUCCESS_RUN_BUILD_SUDO", nil,
			[]string{
				"sudo docker pull node:12",
				fmt.Sprintf("sudo docker container run --rm --entrypoint /bin/sh -e SSH_AUTH_SOCK=/tmp/auth.sock -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd -v %s:/opt/sd/hab -v %s -m2GB --pull never node:12 /opt/sd/local_run.sh ", d.volume, d.habVolume, sshSocket)},
			newBuildEntry(func(b *buildEntry) {
				b.MemoryLimit = "2GB"
			})},
		{"failure build run", "FAIL_BUILD_CONTAINER_RUN_SUDO", fmt.Errorf("failed to run build container: exit status 1"), []string{}, newBuildEntry()},
		{"failure build image pull", "FAIL_BUILD_IMAGE_PULL_SUDO", fmt.Errorf("failed to pull user image exit status 1"), []string{}, newBuildEntry()},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeExecCommand(tt.id)
			execCommand = c.execCmd
			err := d.runBuild(tt.buildEntry)
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

func TestRunBuildWithInteractiveMode(t *testing.T) {
	defer func() {
		execCommand = exec.Command
		osMkdirAll = os.MkdirAll
		osWriteFile = os.WriteFile
	}()

	steps := []screwdriver.Step{
		{Name: "step1", Command: "echo 'This is a test command.'"},
		{Name: "step2", Command: "cat /foo/bar"},
	}

	d := &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        "launcher",
		setupImageVersion: "latest",
		useSudo:           true,
		interactiveMode:   true,
		sdUtilsPath:       ".sd-utils",
		interact:          &mockInteract{},
		socketPath:        os.Getenv("SSH_AUTH_SOCK"),
		dind: DinD{
			enabled:         false,
			volume:          "SD_DIND_CERT",
			shareVolumeName: "SD_DIND_SHARE",
			shareVolumePath: "/opt/sd_dind_share",
			container:       "sd-local-dind",
			network:         "sd-local-dind-bridge",
			image:           "docker:23.0.1-dind-rootless",
		},
	}

	testCase := []struct {
		name             string
		id               string
		expectError      error
		expectedCommands []string
		buildEntry       buildEntry
	}{
		{"success", "SUCCESS_RUN_BUILD_INTERACT", nil,
			[]string{
				"sudo docker pull node:12",
				fmt.Sprintf("sudo docker container run --rm --entrypoint /bin/sh -e SSH_AUTH_SOCK=/tmp/auth.sock -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd -v %s:/opt/sd/hab -v %s -itd -v .sd-utils/:/test/sd-utils --pull never node:12", d.volume, d.habVolume, sshSocket),
				"sudo docker attach "},
			newBuildEntry(func(b *buildEntry) {
				b.Steps = steps
			})},
		{"success with memory limit", "SUCCESS_RUN_BUILD_INTERACT", nil,
			[]string{
				"sudo docker pull node:12",
				fmt.Sprintf("sudo docker container run --rm --entrypoint /bin/sh -e SSH_AUTH_SOCK=/tmp/auth.sock -v /:/sd/workspace/src/screwdriver.cd/sd-local/local-build -v sd-artifacts/:/test/artifacts -v %s:/opt/sd -v %s:/opt/sd/hab -v %s -m2GB -itd -v .sd-utils/:/test/sd-utils --pull never node:12", d.volume, d.habVolume, sshSocket),
				"sudo docker attach SUCCESS_RUN_BUILD_INTERACT"},
			newBuildEntry(func(b *buildEntry) {
				b.MemoryLimit = "2GB"
				b.Steps = steps
			})},
		{"failure build run", "FAIL_BUILD_CONTAINER_RUN_INTERACT", fmt.Errorf("failed to run build container: exit status 1"), []string{}, newBuildEntry(func(b *buildEntry) { b.Steps = steps })},
		{"failure attach build container", "FAIL_BUILD_CONTAINER_ATTACH_INTERACT", fmt.Errorf("failed to attach build container: exit status 1"), []string{}, newBuildEntry(func(b *buildEntry) { b.Steps = steps })},
		{"failure build image pull", "FAIL_BUILD_IMAGE_PULL_INTERACT", fmt.Errorf("failed to pull user image exit status 1"), []string{}, newBuildEntry(func(b *buildEntry) { b.Steps = steps })},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeExecCommand(tt.id)
			fakeOs := newFakeOsMkdir(t)

			execCommand = c.execCmd
			osMkdirAll = fakeOs.mkdirAll
			osWriteFile = fakeOs.WriteFile

			err := d.runBuild(tt.buildEntry)
			for i, expectedCommand := range tt.expectedCommands {
				assert.True(t, strings.Contains(c.commands[i], expectedCommand), "expect %q \nbut got \n%q", expectedCommand, c.commands[i])
			}

			assert.Contains(t, fakeOs.dirPaths, ".sd-utils/bin")
			assert.Contains(t, fakeOs.dirPaths, ".sd-utils/step")
			assert.Contains(t, fakeOs.fileNames, tt.buildEntry.Steps[0].Name)
			assert.Contains(t, fakeOs.fileNames, tt.buildEntry.Steps[1].Name)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDockerKill(t *testing.T) {
	t.Run("success with no commands", func(t *testing.T) {
		defer func() {
			execCommand = exec.Command
			logrus.SetOutput(os.Stderr)
		}()
		d := &docker{
			volume:            "SD_LAUNCH_BIN",
			setupImage:        "launcher",
			setupImageVersion: "latest",
			useSudo:           false,
			mutex:             &sync.Mutex{},
		}
		c := newFakeExecCommand("SUCCESS_TO_KILL")
		execCommand = c.execCmd
		buf := bytes.NewBuffer(nil)
		logrus.SetOutput(buf)
		d.kill(syscall.SIGINT)
		assert.Equal(t, "", buf.String())
	})

	t.Run("success", func(t *testing.T) {
		defer func() {
			execCommand = exec.Command
			logrus.SetOutput(os.Stderr)
		}()
		c := newFakeExecCommand("SUCCESS_TO_KILL")
		execCommand = c.execCmd
		d := &docker{
			volume:            "SD_LAUNCH_BIN",
			setupImage:        "launcher",
			setupImageVersion: "latest",
			useSudo:           false,
			commands:          []*exec.Cmd{execCommand("sleep")},
			mutex:             &sync.Mutex{},
		}

		d.commands[0].Start()
		go func() {
			time.Sleep(waitForKillTime)
			d.mutex.Lock()
			// For some reason, "ProcessState" is not changed in "Process.Signal" or "syscall.kill", so change "ProcessState" directly.
			d.commands[0].ProcessState = &os.ProcessState{}
			d.mutex.Unlock()
		}()

		buf := bytes.NewBuffer(nil)
		logrus.SetOutput(buf)

		d.kill(syscall.SIGINT)

		actual := buf.String()
		assert.Equal(t, "", actual)
	})

	t.Run("failure", func(t *testing.T) {
		defer func() {
			execCommand = exec.Command
			logrus.SetOutput(os.Stderr)
		}()
		c := newFakeExecCommand("FAIL_TO_KILL")
		execCommand = c.execCmd
		command := execCommand("sleep")
		d := &docker{
			volume:            "SD_LAUNCH_BIN",
			setupImage:        "launcher",
			setupImageVersion: "latest",
			useSudo:           false,
			commands:          []*exec.Cmd{command},
			mutex:             &sync.Mutex{},
		}

		d.commands[0].Start()
		PidTmp := d.commands[0].Process.Pid
		defer func() {
			syscall.Kill(PidTmp, syscall.SIGINT)
			signalFn = defaultSignalFunc
		}()

		// Mock the signal function to force an error
		signalFn = func(*os.Process, os.Signal) error {
			return fmt.Errorf("mocked signal failure")
		}

		buf := bytes.NewBuffer(nil)
		logrus.SetOutput(buf)

		d.kill(syscall.SIGINT)

		actual := buf.String()
		expected := "failed to stop process:"
		assert.True(t, strings.Contains(actual, expected), fmt.Sprintf("\nexpected: %s \nactual: %s\n", expected, actual))
	})

	t.Run("success with sudo", func(t *testing.T) {
		defer func() {
			execCommand = exec.Command
		}()
		c := newFakeExecCommand("SUCCESS_TO_KILL")
		execCommand = c.execCmd
		d := &docker{
			volume:            "SD_LAUNCH_BIN",
			setupImage:        "launcher",
			setupImageVersion: "latest",
			useSudo:           true,
			commands:          []*exec.Cmd{execCommand("sleep")},
			mutex:             &sync.Mutex{},
		}

		d.commands[0].Start()
		go func() {
			time.Sleep(waitForKillTime)
			d.mutex.Lock()
			d.commands[0].ProcessState = &os.ProcessState{}
			d.mutex.Unlock()
		}()

		d.kill(syscall.SIGINT)

		assert.Equal(t, fmt.Sprintf("sudo kill -2 %v", d.commands[0].Process.Pid), c.commands[1])
	})
}

func TestDockerClean(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		defer func() {
			execCommand = exec.Command
		}()
		c := newFakeExecCommand("SUCCESS_TO_CLEAN")
		execCommand = c.execCmd
		d := &docker{
			habVolume:         "SD_LAUNCH_HAB",
			volume:            "SD_LAUNCH_BIN",
			setupImage:        "launcher",
			setupImageVersion: "latest",
			commands:          []*exec.Cmd{},
			useSudo:           false,
		}

		d.clean()
		assert.Equal(t, fmt.Sprintf("docker volume rm --force %v", d.habVolume), c.commands[0])
		assert.Equal(t, fmt.Sprintf("docker volume rm --force %v", d.volume), c.commands[1])
	})

	t.Run("success with sudo", func(t *testing.T) {
		defer func() {
			execCommand = exec.Command
		}()
		c := newFakeExecCommand("SUCCESS_TO_CLEAN")
		execCommand = c.execCmd
		d := &docker{
			habVolume:         "SD_LAUNCH_HAB",
			volume:            "SD_LAUNCH_BIN",
			setupImage:        "launcher",
			setupImageVersion: "latest",
			commands:          []*exec.Cmd{},
			useSudo:           true,
		}

		d.clean()
		assert.Equal(t, fmt.Sprintf("sudo docker volume rm --force %v", d.habVolume), c.commands[0])
		assert.Equal(t, fmt.Sprintf("sudo docker volume rm --force %v", d.volume), c.commands[1])
	})

	t.Run("failure", func(t *testing.T) {
		defer func() {
			execCommand = exec.Command
			logrus.SetOutput(os.Stderr)
		}()
		c := newFakeExecCommand("FAIL_TO_CLEAN")
		execCommand = c.execCmd
		d := &docker{
			volume:            "SD_LAUNCH_BIN",
			setupImage:        "launcher",
			setupImageVersion: "latest",
			commands:          []*exec.Cmd{},
			useSudo:           false,
		}

		buf := bytes.NewBuffer(nil)
		logrus.SetOutput(buf)

		d.clean()

		expected := "failed to remove volume:"
		assert.True(t, strings.Contains(buf.String(), expected), fmt.Sprintf("\nexpected: %s \nactual: %s\n", expected, buf.String()))
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

	cmd, subcmd, args := args[0], args[1], args[2:]
	_, _ = cmd, args
	testCase := os.Getenv("GO_TEST_MODE")
	if strings.Index(testCase, "SUDO") > 0 || strings.Index(testCase, "INTERACT") > 0 {
		subcmd = args[0]
	}

	fmt.Print(testCase)

	switch testCase {
	case "":
		os.Exit(1)
	case "SUCCESS_SETUP_BIN":
		os.Exit(0)
	case "SUCCESS_SETUP_BIN_SUDO":
		os.Exit(0)
	case "SUCCESS_SETUP_BIN_INTERACT":
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
	case "SUCCESS_RUN_BUILD_INTERACT":
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
	case "FAIL_BUILD_CONTAINER_RUN_INTERACT":
		if subcmd == "pull" {
			os.Exit(0)
		}
		os.Exit(1)
	case "FAIL_BUILD_CONTAINER_ATTACH_INTERACT":
		if subcmd == "attach" {
			os.Exit(1)
		}
		os.Exit(0)
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
	case "FAIL_BUILD_IMAGE_PULL_INTERACT":
		if subcmd == "pull" {
			os.Exit(1)
		}
		os.Exit(0)
	case "SUCCESS_TO_KILL":
		if subcmd == "sleep" {
			time.Sleep(fakeProcessLifeTime)
			os.Exit(0)
		}
		os.Exit(0)
	case "FAIL_TO_KILL":
		if subcmd == "sleep" {
			time.Sleep(fakeProcessLifeTime)
			os.Exit(0)
		}
		os.Exit(1)
	case "SUCCESS_TO_CLEAN":
		os.Exit(0)
	case "FAIL_TO_CLEAN":
		os.Exit(1)
	}
}
