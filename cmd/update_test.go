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

	cases := []struct {
		name    string
		input   string
		expect1 bool
		expect2 error
	}{
		{
			name:    "Failed input value is n",
			input:   "n",
			expect1: true,
			expect2: nil,
		},
		{
			name:    "Failed input value is N",
			input:   "N",
			expect1: true,
			expect2: nil,
		},
		{
			name:    "Failed input value is no",
			input:   "no",
			expect1: true,
			expect2: nil,
		},
		{
			name:    "Failed input value is No",
			input:   "No",
			expect1: true,
			expect2: nil,
		},
		{
			name:    "Failed input value is y",
			input:   "y",
			expect1: false,
			expect2: nil,
		},
		{
			name:    "Failed input value is Y",
			input:   "Y",
			expect1: false,
			expect2: nil,
		},
		{
			name:    "Failed input value is yes",
			input:   "yes",
			expect1: false,
			expect2: nil,
		},
		{
			name:    "Failed input value is Yes",
			input:   "Yes",
			expect1: false,
			expect2: nil,
		},
		{
			name:    "Failed input value is not n or y",
			input:   "test",
			expect1: false,
			expect2: errors.New("Invalid input"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			aborted, err := checkUserInput(c.input)
			assert.Equal(t, aborted, c.expect1)
			assert.Equal(t, err, c.expect2)
		})
	}
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
