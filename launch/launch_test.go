package launch

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/screwdriver-cd/sd-local/config"
	"github.com/screwdriver-cd/sd-local/screwdriver"
)

var testDir string = "./testdata"

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		buf, _ := ioutil.ReadFile(filepath.Join(testDir, "job.json"))

		job := screwdriver.Job{}

		json.Unmarshal(buf, &job)

		config := config.Config{
			APIURL:   "http://api-test.screwdriver.cd",
			StoreURL: "http://store-test.screwdriver.cd",
			Token:    "testtoken",
			Launcher: config.Launcher{Version: "latest", Image: "screwdrivercd/launcher"},
		}

		expectedBuildConfig := BuildConfig{
			ID: 0,
			Environment: []EnvVar{EnvVar{
				"SD_ARTIFACTS_DIR": "/test/artifacts",
				"SD_API_URL":       "http://api-test.screwdriver.cd",
				"SD_STORE_URL":     "http://store-test.screwdriver.cd",
				"SD_TOKEN":         "testjwt",
				"FOO":              "foo",
			}},
			EventID:       0,
			JobID:         0,
			ParentBuildID: []int{0},
			Sha:           "dummy",
			Meta:          map[string]interface{}{},
			Steps:         job.Steps,
			Image:         job.Image,
			JobName:       "test",
		}

		l := New(job, config, "test", "testjwt")

		assert.Equal(t, expectedBuildConfig, l.buildConfig)
	})
}
