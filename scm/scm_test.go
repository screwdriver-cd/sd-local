package scm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeExecCommand struct {
	id      string
	execCmd func(command string, args ...string) *exec.Cmd
	command string
}

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

		expected := &scm{
			baseDir:   baseDir,
			remoteURL: "https://github.com/screwdriver-cd/sd-local.git",
			branch:    "test",
		}

		s, err := New(baseDir, srcURL)
		defer os.RemoveAll(s.LocalPath())

		expected.localPath = s.LocalPath()
		assert.Equal(t, expected, s)
		assert.Nil(t, err)

		_, err = os.Stat(s.LocalPath())
		assert.Nil(t, err)
	})

	t.Run("success with ssh url", func(t *testing.T) {
		baseDir := os.TempDir()
		srcURL := "git@github.com:screwdriver-cd/sd-local.git#branch#test"

		expected := &scm{
			baseDir:   baseDir,
			remoteURL: "git@github.com:screwdriver-cd/sd-local.git",
			branch:    "branch#test",
		}

		s, err := New(baseDir, srcURL)
		defer os.RemoveAll(filepath.Join(baseDir, "repo"))

		expected.localPath = s.LocalPath()
		assert.Equal(t, expected, s)
		assert.Nil(t, err)

		_, err = os.Stat(s.LocalPath())
		assert.Nil(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		osMkdirAll = func(path string, per os.FileMode) error { return fmt.Errorf("test") }

		baseDir := os.TempDir()
		srcURL := "https://github.com/screwdriver-cd/sd-local.git#test"

		scm, err := New(baseDir, srcURL)
		defer os.RemoveAll(filepath.Join(baseDir, "repo"))

		assert.Nil(t, scm)
		msg := err.Error()
		assert.Equal(t, 0, strings.Index(msg, "failed to make local source directory: "), fmt.Sprintf("expected error is `failed to make local source directory: ...`, actual: `%v`", msg))

		_, err = os.Stat(filepath.Join(baseDir, "repo", "github.com", "screwdriver-cd", "sd-local"))
		assert.NotNil(t, err)
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

		_, err = os.Stat(localPath)

		assert.True(t, os.IsNotExist(err))
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
	}
}
