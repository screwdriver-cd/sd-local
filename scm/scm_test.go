package scm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("success with https url", func(t *testing.T) {
		baseDir := os.TempDir()
		srcUrl := "https://github.com/screwdriver-cd/sd-local.git#test"

		expected := &scm{
			baseDir:   baseDir,
			remoteUrl: "https://github.com/screwdriver-cd/sd-local.git",
			instance:  "github.com",
			org:       "screwdriver-cd",
			repo:      "sd-local",
			branch:    "test",
		}

		scm, err := New(baseDir, srcUrl)
		defer os.RemoveAll(filepath.Join(baseDir, "repo"))

		assert.Equal(t, expected, scm)
		assert.Nil(t, err)

		_, err = os.Stat(filepath.Join(baseDir, "repo", "github.com", "screwdriver-cd", "sd-local"))
		assert.Nil(t, err)
	})

	t.Run("success with ssh url", func(t *testing.T) {
		baseDir := os.TempDir()
		srcUrl := "git@github.com:screwdriver-cd/sd-local.git#branch#test"

		expected := &scm{
			baseDir:   baseDir,
			remoteUrl: "git@github.com:screwdriver-cd/sd-local.git",
			instance:  "github.com",
			org:       "screwdriver-cd",
			repo:      "sd-local",
			branch:    "branch#test",
		}

		scm, err := New(baseDir, srcUrl)
		defer os.RemoveAll(filepath.Join(baseDir, "repo"))

		assert.Equal(t, expected, scm)
		assert.Nil(t, err)

		_, err = os.Stat(filepath.Join(baseDir, "repo", "github.com", "screwdriver-cd", "sd-local"))
		assert.Nil(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		osMkdirAll = func(path string, per os.FileMode) error { return fmt.Errorf("test") }

		baseDir := os.TempDir()
		srcUrl := "https://github.com/screwdriver-cd/sd-local.git#test"

		scm, err := New(baseDir, srcUrl)
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
			baseDir:   "/path/to/base/dir",
			remoteUrl: "https://github.com/screwdriver-cd/sd-local.git",
			instance:  "github.com",
			org:       "screwdriver-cd",
			repo:      "sd-local",
			branch:    "test",
		}

		assert.Equal(
			t,
			"/path/to/base/dir/repo/github.com/screwdriver-cd/sd-local",
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
			remoteUrl: "https://github.com/screwdriver-cd/sd-local.git",
			instance:  "github.com",
			org:       "screwdriver-cd",
			repo:      "sd-local",
			branch:    "test",
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
