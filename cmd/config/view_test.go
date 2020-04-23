package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testConfig string

func TestViewCmd(t *testing.T) {
	fp := filePath
	defer func() {
		filePath = fp
	}()

	filePath = func() (string, error) {
		return testConfig, nil
	}

	testCase := []struct {
		name   string
		args   []string
		expect []string
		config string
	}{
		{
			name: "success",
			args: []string{"view"},
			expect: []string{`* default:
    api-url: api.screwdriver.com
    store-url: store.screwdriver.com
    token: sd-token
    launcher:
      version: 1.0.0
      image: screwdrivercd/launcher`,
				`  test:
    api-url: api-test.screwdriver.com
    store-url: store-test.screwdriver.com
    token: sd-token-test
    launcher:
      version: 1.0.0-test
      image: screwdrivercd/launcher
`},
			config: "./testdata/config",
		},
		{
			name: "success with no current",
			args: []string{"view"},
			expect: []string{`  default:
    api-url: api.screwdriver.com
    store-url: store.screwdriver.com
    token: sd-token
    launcher:
      version: 1.0.0
      image: screwdrivercd/launcher`,
				`  test:
    api-url: api-test.screwdriver.com
    store-url: store-test.screwdriver.com
    token: sd-token-test
    launcher:
      version: 1.0.0-test
      image: screwdrivercd/launcher
`},
			config: "./testdata/config_no_current",
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			testConfig = tt.config
			cmd := NewConfigCmd()
			cmd.SetArgs(tt.args)
			buf := bytes.NewBuffer(nil)
			cmd.SetOut(buf)
			err := cmd.Execute()
			if err != nil {
				t.Fatal(err)
			}
			actual := buf.String()
			for _, expect := range tt.expect {
				// we do not use assert.Equal but string.Contains.
				// because maps deserialized by go-yaml has an order different from the written order.
				assert.True(t, strings.Contains(actual, expect), "expect to contain %q \nbut got \n%q", expect, actual)

			}
		})
	}
}
