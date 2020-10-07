package cmd

import (
	"bytes"
	"testing"

	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/stretchr/testify/assert"
)

func setVersion(v string) {
	currentVersion = v
}

func TestSelfUpdate(t *testing.T) {

	t.Run("Failed sd-local version is dev", func(t *testing.T) {
		cmd := newUpdateCmd()
		setVersion("dev")
		updateFlag = true
		buf := bytes.NewBuffer(nil)
		cmd.SetOut(buf)
		cmd.Execute()
		want := "Error: This is a development  and cannot be updated.\nUsage:\n  update [flags]\n\nFlags:\n  -h, --help   help for update\n  -y, --yes    answer yes for all questions\n\n"
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

	t.Run("Failed wrong repository name", func(t *testing.T) {
		cmd := newUpdateCmd()
		setVersion("1.0.4")
		b_githubSlug := githubSlug
		githubSlug = "screwdriver-cd/sd-local-test"
		defer func() { githubSlug = b_githubSlug }()
		updateFlag = true
		buf := bytes.NewBuffer(nil)
		cmd.SetOut(buf)
		cmd.Execute()
		want := "Error: Repositry Not Found\nUsage:\n  update [flags]\n\nFlags:\n  -h, --help   help for update\n  -y, --yes    answer yes for all questions\n\n"
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
