package cmd

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsAborted(t *testing.T) {

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
			expect1: true,
			expect2: errors.New("Invalid input"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			aborted, err := isAborted(c.input)
			assert.Equal(t, aborted, c.expect1)
			assert.Equal(t, err, c.expect2)
		})
	}
}

func TestSelfUpdate(t *testing.T) {
	defaultDetectLatest := detectLatest
	defaultUpdateTo := updateTo
	defer func() {
		detectLatest = defaultDetectLatest
		updateTo = defaultUpdateTo
	}()

	detectLatest = func(slug string) (*selfupdate.Release, bool, error) {
		latest := semver.Version{
			Major: 1,
			Minor: 0,
			Patch: 5,
		}
		return &selfupdate.Release{Version: latest}, true, nil
	}
	updateTo = func(url, path string) error {
		return nil
	}

	testCases := []struct {
		name      string
		current   string
		errOutput string
		logOutput []string
	}{
		{
			name:      "Failed current version is dev",
			current:   "dev",
			errOutput: "This is a development version and cannot be updated",
			logOutput: []string{
				"Current version: dev",
			},
		},
		{
			name:      "Failed current version is latest",
			current:   "1.0.5",
			errOutput: "",
			logOutput: []string{
				"Current version: 1.0.5",
				"Current version is latest",
			},
		},
		{
			name:      "Success selfUpdate command",
			current:   "1.0.4",
			errOutput: "",
			logOutput: []string{
				"Current version: 1.0.4",
				"Updating ...",
				"Successfully updated to version 1.0.5",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			backupVersion := version
			defer func() {
				version = backupVersion
				logrus.SetOutput(os.Stderr)
			}()

			version = tt.current
			logBuf := bytes.NewBuffer(nil)
			logrus.SetOutput(logBuf)

			cmd := newUpdateCmd()
			updateFlag = true
			errBuf := bytes.NewBuffer(nil)
			cmd.SetOut(errBuf)
			cmdErr := cmd.Execute()

			if tt.errOutput != "" {
				assert.Equal(t, tt.errOutput, cmdErr.Error())

			}

			for _, want := range tt.logOutput {
				assert.Contains(t, logBuf.String(), want)
			}
		})
	}
}
