package launch

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type docker struct {
	volume            string
	setupImage        string
	setupImageVersion string
}

var _ Runner = (*docker)(nil)

func newDocker(setupImage, setupImageVer string) Runner {
	return &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        setupImage,
		setupImageVersion: setupImageVer,
	}
}

func (d *docker) SetupBin() error {
	err := exec.Command("sudo", "docker", "volume", "create", "--name", d.volume).Run()
	if err != nil {
		return err
	}

	mount := fmt.Sprintf("%s:/opt/sd/", d.volume)
	image := fmt.Sprintf("%s:%s", d.setupImage, d.setupImageVersion)
	cmd := exec.Command("sudo", "docker", "run", "-v", mount, image, "--entrypoint", "/bin/echo set up bin")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func (d *docker) RunBuild(buildConfig BuildConfig, environment BuildEnvironment) ([]byte, error) {
	//厳密にするならカレントかつscrewdriver.yamlがある場所にした方が良さそう
	cwd, err := os.Getwd()

	if err != nil {
		return nil, nil
	}

	srcDir := cwd
	hostArtDir := cwd
	containerArtDir := environment.SD_ARTIFACTS_DIR
	buildImage := buildConfig.Image

	srcOpt := fmt.Sprintf("%s/:/sd/workspace", srcDir)
	artOpt := fmt.Sprintf("%s/:%s", hostArtDir, containerArtDir)
	binOpt := fmt.Sprintf("%s:%s", d.volume, "/opt/sd")
	configJson, err := json.Marshal(buildConfig)
	if err != nil {
		return nil, err
	}

	cmd := []string{"docker", "run", "--rm", "-v", srcOpt, "-v", artOpt, "-v", binOpt, buildImage, "/opt/sd/local_run.sh", string(configJson), buildConfig.JobName, environment.SD_API_URL, environment.SD_STORE_URL, containerArtDir}
	out, err := exec.Command("sudo", cmd...).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return nil, err
	}

	return out, nil
}
