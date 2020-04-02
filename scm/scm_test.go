package scm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

type fakeExecCommand struct {
	id      string
	execCmd func(command string, args ...string) *exec.Cmd
	command string
}

const (
	fakeProcessLifeTime = 100 * time.Second
)

func newFakeExecCommand(id string) *fakeExecCommand {
	c := &fakeExecCommand{}
	c.id = id
	c.execCmd = func(name string, args ...string) *exec.Cmd {
		c.command = fmt.Sprintf("%s %s", name, strings.Join(args, " "))
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_TEST_MODE=%s", id)}
		return cmd
	}
	return c
}

func TestNew(t *testing.T) {
	t.Run("success with https url", func(t *testing.T) {
		baseDir := os.TempDir()
		srcURL := "https://github.com/screwdriver-cd/sd-local.git#test"

		s, err := New(baseDir, srcURL, false)
		defer os.RemoveAll(s.LocalPath())

		scm := s.(*scm)

		assert.Equal(t, baseDir, scm.baseDir)
		assert.Equal(t, "https://github.com/screwdriver-cd/sd-local.git", scm.remoteURL)
		assert.Equal(t, "test", scm.branch)
		assert.NotEmpty(t, scm.LocalPath())
		assert.DirExists(t, scm.LocalPath())
		assert.Nil(t, err)
	})

	t.Run("success with ssh url", func(t *testing.T) {
		baseDir := os.TempDir()
		srcURL := "git@github.com:screwdriver-cd/sd-local.git#branch#test"

		s, err := New(baseDir, srcURL, false)
		defer os.RemoveAll(s.LocalPath())

		scm := s.(*scm)

		assert.Equal(t, baseDir, scm.baseDir)
		assert.Equal(t, "git@github.com:screwdriver-cd/sd-local.git", scm.remoteURL)
		assert.Equal(t, "branch#test", scm.branch)
		assert.NotEmpty(t, scm.LocalPath())
		assert.DirExists(t, scm.LocalPath())
		assert.Nil(t, err)
	})

	t.Run("failure with making directory", func(t *testing.T) {
		osMkdirAll = func(path string, per os.FileMode) error { return fmt.Errorf("test") }

		baseDir := os.TempDir()
		srcURL := "https://github.com/screwdriver-cd/sd-local.git#test"

		s, err := New(baseDir, srcURL, false)
		msg := err.Error()

		assert.Nil(t, s)
		assert.Equal(t, 0, strings.Index(msg, "failed to make local source directory: "), fmt.Sprintf("expected error is `failed to make local source directory: ...`, actual: `%v`", msg))
	})

	t.Run("failure with invalid url", func(t *testing.T) {
		osMkdirAll = func(path string, per os.FileMode) error { return fmt.Errorf("test") }

		baseDir := os.TempDir()
		srcURL := "https://github.com/screwdriver-cd"

		s, err := New(baseDir, srcURL, false)

		assert.Nil(t, s)
		assert.Equal(t, err.Error(), "failed to fetch source code with invalid URL: https://github.com/screwdriver-cd")
	})
}

func TestLocalPath(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := &scm{
			localPath: "/path/to/base/dir/repo/foobar",
		}

		assert.Equal(
			t,
			"/path/to/base/dir/repo/foobar",
			s.LocalPath(),
		)

	})
}

func TestKill(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		defer func() {
			execCommand = exec.Command
			logrus.SetOutput(os.Stderr)
		}()
		c := newFakeExecCommand("SUCCESS_TO_KILL")
		execCommand = c.execCmd
		command := execCommand("sleep")
		s := &scm{
			commands: []*exec.Cmd{command},
		}

		s.commands[0].Start()

		buf := bytes.NewBuffer(nil)
		logrus.SetOutput(buf)

		s.Kill(syscall.SIGINT)

		actual := buf.String()
		assert.Equal(t, "", actual)
	})

	t.Run("failure", func(t *testing.T) {
		defer func() {
			execCommand = exec.Command
			logrus.SetOutput(os.Stderr)
		}()
		c := newFakeExecCommand("FAILED_TO_KILL")
		execCommand = c.execCmd
		command := execCommand("sleep")
		s := &scm{
			commands: []*exec.Cmd{command},
		}

		s.commands[0].Start()

		PidTmp := s.commands[0].Process.Pid
		defer func() {
			syscall.Kill(PidTmp, syscall.SIGINT)
		}()
		s.commands[0].Process.Pid = 0

		buf := bytes.NewBuffer(nil)
		logrus.SetOutput(buf)

		s.Kill(syscall.SIGINT)

		actual := buf.String()
		expected := "failed to stop process:"
		assert.True(t, strings.Contains(actual, expected), fmt.Sprintf("\nexpected: %s \nactual: %s\n", expected, actual))
	})
}

func TestClean(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		baseDir := os.TempDir()
		defer os.RemoveAll(filepath.Join(baseDir, "repo"))
		s := &scm{
			baseDir:   baseDir,
			remoteURL: "https://github.com/screwdriver-cd/sd-local.git",
			branch:    "test",
			localPath: filepath.Join(baseDir, "repo/test"),
		}

		localPath := s.LocalPath()

		err := os.MkdirAll(localPath, 0777)

		if err != nil {
			t.Fatal(err)
		}

		s.Clean()

		assert.NoDirExists(t, s.LocalPath())
	})

	t.Run("success with sudo", func(t *testing.T) {
		baseDir := os.TempDir()
		defer func() {
			os.RemoveAll(filepath.Join(baseDir, "repo"))
			logrus.SetOutput(os.Stderr)
		}()

		c := newFakeExecCommand("SUCCESS_TO_CLEAN")
		execCommand = c.execCmd
		s := &scm{
			baseDir:   baseDir,
			remoteURL: "https://github.com/screwdriver-cd/sd-local.git",
			branch:    "test",
			localPath: filepath.Join(baseDir, "repo/test"),
			sudo:      true,
		}

		buf := bytes.NewBuffer(nil)
		logrus.SetOutput(buf)

		s.Clean()

		assert.Equal(t, "", buf.String())
	})

	t.Run("failure", func(t *testing.T) {
		baseDir := os.TempDir()
		defer func() {
			os.RemoveAll(filepath.Join(baseDir, "repo"))
			logrus.SetOutput(os.Stderr)
		}()

		c := newFakeExecCommand("FAILURE_TO_CLEAN")
		execCommand = c.execCmd
		s := &scm{
			baseDir:   baseDir,
			remoteURL: "https://github.com/screwdriver-cd/sd-local.git",
			branch:    "test",
			localPath: filepath.Join(baseDir, "repo/test"),
			sudo:      true,
		}

		buf := bytes.NewBuffer(nil)
		logrus.SetOutput(buf)

		s.Clean()

		actual := buf.String()
		expected := "failed to remove local source directory:"
		assert.True(t, strings.Contains(actual, expected), fmt.Sprintf("\nexpected: %s\nactual: %s\n", expected, actual))
	})
}

func TestPull(t *testing.T) {
	t.Run("success with branch", func(t *testing.T) {
		baseDir := os.TempDir()
		defer os.RemoveAll(filepath.Join(baseDir, "repo"))
		s := &scm{
			baseDir:   baseDir,
			remoteURL: "https://github.com/screwdriver-cd/sd-local.git",
			branch:    "test",
			localPath: filepath.Join(baseDir, "repo/test"),
		}
		c := newFakeExecCommand("SUCCESS_PULL")
		execCommand = c.execCmd
		os.MkdirAll(s.LocalPath(), 0777)

		err := s.Pull()
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("git clone -b test https://github.com/screwdriver-cd/sd-local.git %s", s.LocalPath()), c.command)
	})

	t.Run("success without branch", func(t *testing.T) {
		baseDir := os.TempDir()
		defer os.RemoveAll(filepath.Join(baseDir, "repo"))
		s := &scm{
			baseDir:   baseDir,
			remoteURL: "https://github.com/screwdriver-cd/sd-local.git",
			localPath: filepath.Join(baseDir, "repo/test"),
		}
		c := newFakeExecCommand("SUCCESS_PULL")
		execCommand = c.execCmd
		os.MkdirAll(s.LocalPath(), 0777)

		err := s.Pull()
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("git clone https://github.com/screwdriver-cd/sd-local.git %s", s.LocalPath()), c.command)
	})

	t.Run("failed to pull image", func(t *testing.T) {
		s := &scm{
			remoteURL: "https://github.com/screwdriver-cd/sd-local.git",
		}
		c := newFakeExecCommand("FAILED_PULL")
		execCommand = c.execCmd
		os.MkdirAll(s.LocalPath(), 0777)

		err := s.Pull()
		assert.NotNil(t, err)
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
	_, _, _ = cmd, subcmd, args

	switch os.Getenv("GO_TEST_MODE") {
	case "":
		os.Exit(1)
	case "SUCCESS_PULL":
		os.Exit(0)
	case "FAILED_PULL":
		os.Exit(1)
	case "SUCCESS_TO_KILL":
		if subcmd == "sleep" {
			time.Sleep(fakeProcessLifeTime)
			os.Exit(0)
		}
	case "FAIL_TO_KILL":
		if subcmd == "sleep" {
			time.Sleep(fakeProcessLifeTime)
			os.Exit(0)
		}
		os.Exit(1)
	case "SUCCESS_TO_CLEAN":
		os.Exit(0)
	case "FAILURE_TO_CLEAN":
		os.Exit(1)
	}
}
