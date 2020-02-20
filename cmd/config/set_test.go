package config

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/screwdriver-cd/sd-local/config"

	"github.com/stretchr/testify/assert"
)

func TestConfigSetCmd(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	cnfPath := fmt.Sprintf("%vconfig", rand.Int())
	defer os.Remove(cnfPath)

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
			args:     []string{"set", "api-url", "example.com"},
			wantOut:  "",
			checkErr: false,
		},
		{
			name:     "failure by too many args",
			args:     []string{"set", "api-url", "example.com", "many"},
			wantOut:  "",
			checkErr: true,
		},
		{
			name:     "failure by too little args",
			args:     []string{"set", "api-url"},
			wantOut:  "",
			checkErr: true,
		},
		{
			name:     "success by use option",
			args:     []string{"set", "api-url", "example.com", "--local"},
			wantOut:  "",
			checkErr: false,
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
