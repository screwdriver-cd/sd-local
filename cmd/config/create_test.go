package config

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/screwdriver-cd/sd-local/config"

	"github.com/stretchr/testify/assert"
)

func createRandNameConfig(conf io.Reader) (string, error) {
	rand.Seed(time.Now().UnixNano())
	cnfPath := fmt.Sprintf("%vconfig", rand.Int())
	f, err := os.Create(cnfPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, conf)
	if err != nil {
		os.Remove(cnfPath)
		return "", err
	}
	return cnfPath, nil
}
func TestConfigCreateCmd(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	cnfPath := fmt.Sprintf("%vconfig", rand.Int())
	defer os.Remove(cnfPath)

	cnew := configNew
	defer func() {
		configNew = cnew
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
			args:     []string{"create", "test"},
			wantOut:  "",
			checkErr: false,
		},
		{
			name:     "failure by Entry that already exists",
			args:     []string{"create", "default"},
			wantOut:  "",
			checkErr: true,
		},
		{
			name:     "failure by too many args",
			args:     []string{"create", "test", "many"},
			wantOut:  "",
			checkErr: true,
		},
		{
			name:     "failure by too little args",
			args:     []string{"create"},
			wantOut:  "",
			checkErr: true,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewConfigCmd()
			cmd.SetArgs(tt.args)
			buf := bytes.NewBuffer(nil)
			cmd.SetOut(buf)
			err := cmd.Execute()
			if tt.checkErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantOut, buf.String())
			}

		})
	}

}
