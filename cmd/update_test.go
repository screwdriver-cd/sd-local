package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/stretchr/testify/assert"
)

func setVersion(v string) {
	currentVersion = v
}

func TestCheckUserInput(t *testing.T) {

	t.Run("Failed input value is n", func(t *testing.T) {
		err := checkUserInput("n")
		want := errors.New("Aborted")
		assert.Equal(t, want, err)
	})
	t.Run("Failed input value is N", func(t *testing.T) {
		err := checkUserInput("N")
		want := errors.New("Aborted")
		assert.Equal(t, want, err)
	})
	t.Run("Failed input value is not n or y", func(t *testing.T) {
		err := checkUserInput("test")
		want := errors.New("Invalid input")
		assert.Equal(t, want, err)
	})
	t.Run("Success input value is y", func(t *testing.T) {
		err := checkUserInput("y")
		assert.Equal(t, nil, err)
	})
	t.Run("Success input value is Y", func(t *testing.T) {
		err := checkUserInput("Y")
		assert.Equal(t, nil, err)
	})
}

func TestSelfUpdate(t *testing.T) {

	t.Run("Failed current version is dev", func(t *testing.T) {
		cmd := newUpdateCmd()
		setVersion("dev")
		updateFlag = true
		buf := bytes.NewBuffer(nil)
		cmd.SetOut(buf)
		cmd.Execute()
		want := "Error: This is a development version and cannot be updated\nUsage:\n  update [flags]\n\nFlags:\n  -h, --help   help for update\n  -y, --yes    answer yes for all questions\n\n"
		assert.Equal(t, want, buf.String())
	})

	t.Run("Failed current version is latest", func(t *testing.T) {
		cmd := newUpdateCmd()
		latest, _, _ := selfupdate.DetectLatest(githubSlug)
		setVersion(latest.Version.String())
		updateFlag = true
		buf := bytes.NewBuffer(nil)
		cmd.SetOut(buf)
		cmd.Execute()
		want := "Error: Current version is latest\nUsage:\n  update [flags]\n\nFlags:\n  -h, --help   help for update\n  -y, --yes    answer yes for all questions\n\n"
		assert.Equal(t, want, buf.String())
	})

	t.Run("Success selfUpdate command", func(t *testing.T) {
		cmd := newUpdateCmd()
		setVersion("1.0.4")
		updateFlag = true
		buf := bytes.NewBuffer(nil)
		cmd.SetOut(buf)
		err := cmd.Execute()
		want := ""
		assert.Equal(t, want, buf.String())
		assert.Nil(t, err)
	})
}
