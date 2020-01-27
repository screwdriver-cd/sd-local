package launch

import (
	"fmt"
	"github.com/screwdriver-cd/sd-local/screwdriver"
)

type EnvVar map[string]string


type LaunchConfig struct {
	ID            int                    `json:"id"`
	Environment   []EnvVar               `json:"environment"`
	EventID       int                    `json:"eventId"`
	JobID         int                    `json:"jobId"`
	ParentBuildID []int                  `json:"parentBuildId"`
	Sha           string                 `json:"sha"`
	Meta          map[string]interface{} `json:"meta"`
	Steps         []screwdriver.Step                 `json:"steps"`
}

type deepCopyEnvVar []EnvVar

// EnvVar配列同士をマージする
func envMerge(env1 deepCopyEnvVar, env2 deepCopyEnvVar) deepCopyEnvVar{
	merged := make([]EnvVar, len(env1))

	for k, v := range env2 {
		merged = append(merged, EnvVar{k: v})
	}

	return merged
}

// EnvVar配列をDeepCopyしてEnvVar配列を返す関数 => EnvFunc1 

// EnvVarからDeepCopyのEnvVar配列を返す関数 => EnvFunc2


func CreateLaunchConfig(job screwdriver.Job, token string) LaunchConfig {
	t1 := []EnvVar{EnvVar{"SD_TOKEN": token}}
	env := envMerge(t1, job.Environment)
	env[0]["HOGE"] = "HOGE"
	fmt.Println(t1)

	return LaunchConfig{
		ID: 0,
		Environment: env,
		EventID: 0,
		JobID: 0,
		ParentBuildID: []int{0},
		Sha: "dummy",
		Meta: map[string]interface{}{},
		Steps: job.Steps,
	}
}