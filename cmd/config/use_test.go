package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/screwdriver-cd/sd-local/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigUseCmd(t *testing.T) {
	f, err := os.Open("./testdata/config")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	cnfPath, err := createRandNameConfig(f)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cnfPath)

	preconf := configNew
	defer func() {
		configNew = preconf
	}()
	configNew = func(configPath string) (c config.Config, err error) {
		return config.New(cnfPath)
	}

	testCase := []struct {
		name     string
		args     []string
		wantOut  string
		checkErr bool
	}{
		{
			name:     "success",
			args:     []string{"use", "test"},
			wantOut:  "",
			checkErr: false,
		},
		{
			name:     "failure with too many args",
			args:     []string{"use", "test", "args"},
			wantOut:  "accepts 1 arg(s), received 2",
			checkErr: true,
		},
		{name: "failure without args",
			args:     []string{"use"},
			wantOut:  "accepts 1 arg(s), received 0",
			checkErr: true,
		},
		{
			name:     "failure because of passing unknown config",
			args:     []string{"use", "unknownconfig"},
			wantOut:  "config `unknownconfig` does not exist",
			checkErr: true,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewConfigCmd()
			cmd.SilenceUsage = true
			cmd.SetArgs(tt.args)
			buf := bytes.NewBuffer(nil)
			cmd.SetOut(buf)
			err := cmd.Execute()
			if tt.checkErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.wantOut, err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantOut, buf.String())
			}

		})
	}
}
