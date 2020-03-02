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
	_ = execCommand("docker", "volume", "rm", "--force", d.volume).Run()

	err := execCommand("docker", "volume", "create", "--name", d.volume).Run()
	if err != nil {
		return fmt.Errorf("failed to create docker volume")
	}

	mount := fmt.Sprintf("%s:/opt/sd/", d.volume)
	image := fmt.Sprintf("%s:%s", d.setupImage, d.setupImageVersion)
	err = execCommand("docker", "pull", image).Run()
	if err != nil {
		return fmt.Errorf("failed to pull launcher image %v", err)
	}
	cmd := execCommand("docker", "container", "run", "--rm", "-v", mount, image, "--entrypoint", "/bin/echo set up bin")
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		io.Copy(os.Stderr, buf)
		return fmt.Errorf("failed to prepare build scripts")
	}

	return nil
}

func (d *docker) runBuild(buildConfig buildConfig) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	environment := buildConfig.Environment[0]

	srcDir := cwd
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

	err = execCommand("docker", "pull", buildImage).Run()
	if err != nil {
		return fmt.Errorf("failed to pull user image %v", err)
	}
	cmd := execCommand("docker", "container", "run", "--rm", "-v", srcVol, "-v", artVol, "-v", binVol, buildImage, "/opt/sd/local_run.sh", string(configJSON), buildConfig.JobName, environment["SD_API_URL"], environment["SD_STORE_URL"], logfilePath)

	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		io.Copy(os.Stderr, buf)
		return err
	}

	return nil
}
