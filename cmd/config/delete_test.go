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

func TestConfigDeleteCmd(t *testing.T) {
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

	f, err := os.Create(cnfPath)
	if err != nil {
		t.Fatal(err)
	}
	cf, err := os.Open("./testdata/config")
	_, err = io.Copy(f, cf)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = cf.Close()
	if err != nil {
		t.Fatal(err)
	}

	testCase := []struct {
		name     string
		args     []string
		wantOut  string
		checkErr bool
	}{
		{
			name:     "success",
			args:     []string{"delete", "test"},
			wantOut:  "",
			checkErr: false,
		},
		{
			name:     "failure by Entry that does not exist",
			args:     []string{"delete", "test"},
			wantOut:  "",
			checkErr: true,
		},
		{
			name:     "failure by too many args",
			args:     []string{"delete", "test", "many"},
			wantOut:  "",
			checkErr: true,
		},
		{
			name:     "failure by too little args",
			args:     []string{"delete"},
			wantOut:  "",
			checkErr: true,
		},
		{
			name:     "failure by trying delete current config",
			args:     []string{"delete", "default"},
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
