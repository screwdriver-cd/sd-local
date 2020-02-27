package launch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type docker struct {
	volume            string
	setupImage        string
	setupImageVersion string
}

var _ runner = (*docker)(nil)
var execCommand = exec.Command

const (
	// ArtifactsDir is default artifact directory name
	ArtifactsDir = "sd-artifacts"
	// LogFile is default logfile name for build log
	LogFile = "builds.log"
	// The definition of "ScmHost" and "OrgRepo" is in "PipelineFromID" of "screwdriver/screwdriver_local.go"
	scmHost = "screwdriver.cd"
	orgRepo = "sd-local/local-build"
)

func newDocker(setupImage, setupImageVer string) runner {
	return &docker{
		volume:            "SD_LAUNCH_BIN",
		setupImage:        setupImage,
		setupImageVersion: setupImageVer,
	}
}

func (d *docker) setupBin() error {
	_ = execDockerCommand("volume", "rm", "--force", d.volume)

	err := execDockerCommand("volume", "create", "--name", d.volume)
	if err != nil {
		return fmt.Errorf("failed to create docker volume: %v", err)
	}

	mount := fmt.Sprintf("%s:/opt/sd/", d.volume)
	image := fmt.Sprintf("%s:%s", d.setupImage, d.setupImageVersion)
	err = execDockerCommand("pull", image)
	if err != nil {
		return fmt.Errorf("failed to pull launcher image: %v", err)
	}

	err = execDockerCommand("container", "run", "--rm", "-v", mount, image, "--entrypoint", "/bin/echo set up bin")
	if err != nil {
		return fmt.Errorf("failed to prepare build scripts: %v", err)
	}

	return nil
}

func (d *docker) runBuild(buildConfig buildConfig) error {
	environment := buildConfig.Environment[0]

	srcDir := buildConfig.srcPath
	hostArtDir := buildConfig.ArtifactsPath
	containerArtDir := environment["SD_ARTIFACTS_DIR"]
	buildImage := buildConfig.Image
	logfilePath := filepath.Join(containerArtDir, LogFile)

	srcVol := fmt.Sprintf("%s/:/sd/workspace/src/%s/%s", srcDir, scmHost, orgRepo)
	artVol := fmt.Sprintf("%s/:%s", hostArtDir, containerArtDir)
	binVol := fmt.Sprintf("%s:%s", d.volume, "/opt/sd")
	configJSON, err := json.Marshal(buildConfig)
	if err != nil {
		return err
	}

	err = execDockerCommand("pull", buildImage)
	if err != nil {
		return fmt.Errorf("failed to pull user image %v", err)
	}

	err = execDockerCommand("container", "run", "--rm", "-v", srcVol, "-v", artVol, "-v", binVol, buildImage, "/opt/sd/local_run.sh", string(configJSON), buildConfig.JobName, environment["SD_API_URL"], environment["SD_STORE_URL"], logfilePath)
	if err != nil {
		return fmt.Errorf("failed to run build container: %v", err)
	}

	return nil
}

func execDockerCommand(args ...string) error {
	cmd := execCommand("docker", args...)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		io.Copy(os.Stderr, buf)
		return err
	}
	return nil
}
