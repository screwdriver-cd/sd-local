package config

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigSetCmd(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	cnfPath := fmt.Sprintf("%vconfig", rand.Int())
	defer os.Remove(cnfPath)

	defFilePath := filePath
	defer func() {
		filePath = defFilePath
	}()
	filePath = func() (string, error) {
		return cnfPath, nil
	}

	testCase := []struct {
		name     string
		args     []string
		existErr bool
	}{
		{
			name:     "success",
			args:     []string{"set", "api-url", "example.com"},
			existErr: false,
		},
		{
			name:     "failure by too many args",
			args:     []string{"set", "api-url", "example.com", "many"},
			existErr: true,
		},
		{
			name:     "failure by too little args",
			args:     []string{"set", "api-url"},
			existErr: true,
		},
		{
			name:     "failure by setting an invalid key",
			args:     []string{"set", "invalid-key", "invalid-value"},
			existErr: true,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewConfigCmd()
			cmd.SetArgs(tt.args)
			buf := bytes.NewBuffer(nil)
			cmd.SetOut(buf)
			err := cmd.Execute()
			assert.Equal(t, tt.existErr, err != nil)
		})
	}
}
