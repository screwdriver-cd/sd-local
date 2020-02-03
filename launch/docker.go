package launch

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

type docker struct {
	volume            string
	setupImage        string
	setupImageVersion string
}

var _ Runner = (*docker)(nil)
var execCommand = exec.Command

const (
	artifactsDir = "artifacts"
	logFile      = "builds.log"
)

func newDocker(setupImage, setupImageVer string) Runner {
	return &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        setupImage,
		setupImageVersion: setupImageVer,
	}
}

func (d *docker) SetupBin() error {
	err := execCommand("sudo", "docker", "volume", "create", "--name", d.volume).Run()
	if err != nil {
		return fmt.Errorf("failed to create docker volume")
	}

	mount := fmt.Sprintf("%s:/opt/sd/", d.volume)
	image := fmt.Sprintf("%s:%s", d.setupImage, d.setupImageVersion)
	cmd := execCommand("sudo", "docker", "container", "run", "-v", mount, image, "--entrypoint", "/bin/echo set up bin")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		return fmt.Errorf("failed to prepare build scripts")
	}

	return nil
}

func (d *docker) RunBuild(buildConfig BuildConfig) ([]byte, error) {
	//厳密にするならカレントかつscrewdriver.yamlがある場所にした方が良さそう
	cwd, err := os.Getwd()

	if err != nil {
		return nil, nil
	}

	environment := buildConfig.Environment[0]

	srcDir := cwd
	hostArtDir := filepath.Join(cwd, artifactsDir)
	containerArtDir := environment["SD_ARTIFACTS_DIR"]
	buildImage := buildConfig.Image
	logfilePath := path.Join(containerArtDir, logFile)

	srcVol := fmt.Sprintf("%s/:/sd/workspace", srcDir)
	artVol := fmt.Sprintf("%s/:%s", hostArtDir, containerArtDir)
	binVol := fmt.Sprintf("%s:%s", d.volume, "/opt/sd")
	configJson, err := json.Marshal(buildConfig)
	if err != nil {
		return nil, err
	}

	cmd := []string{"docker", "container", "run", "--rm", "-v", srcVol, "-v", artVol, "-v", binVol, buildImage, "/opt/sd/local_run.sh", string(configJson), buildConfig.JobName, environment["SD_API_URL"], environment["SD_STORE_URL"], logfilePath}
	out, err := execCommand("sudo", cmd...).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return nil, err
	}

	return out, nil
}
