package launch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/screwdriver-cd/sd-local/screwdriver"
	"github.com/sirupsen/logrus"
)

type docker struct {
	volume            string
	habVolume         string
	setupImage        string
	setupImageVersion string
	useSudo           bool
	interactMode      bool
	commands          []*exec.Cmd
	mutex             *sync.Mutex
	flagVerbose       bool
	interact          Interact
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

func newDocker(setupImage, setupImageVer string, useSudo bool, interactMode bool, flagVerbose bool) runner {
	return &docker{
		volume:            "SD_LAUNCH_BIN",
		habVolume:         "SD_LAUNCH_HAB",
		setupImage:        setupImage,
		setupImageVersion: setupImageVer,
		useSudo:           useSudo,
		interactMode:      interactMode,
		commands:          make([]*exec.Cmd, 0, 10),
		mutex:             &sync.Mutex{},
		flagVerbose:       flagVerbose,
		interact:          &InteractImpl{},
	}
}

func (d *docker) setupBin() error {
	_, err := d.execDockerCommand("volume", "create", "--name", d.volume)
	if err != nil {
		return fmt.Errorf("failed to create docker volume: %v", err)
	}

	_, err = d.execDockerCommand("volume", "create", "--name", d.habVolume)
	if err != nil {
		return fmt.Errorf("failed to create docker hab volume: %v", err)
	}

	mount := fmt.Sprintf("%s:/opt/sd/", d.volume)
	habMount := fmt.Sprintf("%s:/hab", d.habVolume)
	image := fmt.Sprintf("%s:%s", d.setupImage, d.setupImageVersion)
	_, err = d.execDockerCommand("pull", image)
	if err != nil {
		return fmt.Errorf("failed to pull launcher image: %v", err)
	}

	_, err = d.execDockerCommand("container", "run", "--rm", "-v", mount, "-v", habMount, image, "--entrypoint", "/bin/echo set up bin")
	if err != nil {
		return fmt.Errorf("failed to prepare build scripts: %v", err)
	}

	return nil
}

func (d *docker) runBuild(buildEntry buildEntry) error {
	environment := buildEntry.Environment[0]

	srcDir := buildEntry.SrcPath
	hostArtDir := buildEntry.ArtifactsPath
	containerArtDir := environment["SD_ARTIFACTS_DIR"]
	buildImage := buildEntry.Image
	logfilePath := filepath.Join(containerArtDir, LogFile)

	srcVol := fmt.Sprintf("%s/:/sd/workspace/src/%s/%s", srcDir, scmHost, orgRepo)
	artVol := fmt.Sprintf("%s/:%s", hostArtDir, containerArtDir)
	binVol := fmt.Sprintf("%s:%s", d.volume, "/opt/sd")
	habVol := fmt.Sprintf("%s:%s", d.habVolume, "/opt/sd/hab")

	// Overwrite steps for sd-local interact mode. The env will load later.
	if d.interactMode {
		buildEntry.Steps = []screwdriver.Step{
			{
				Name:    "sd-local-init",
				Command: "env > /tmp/sd-local.env",
			},
		}
	}

	configJSON, err := json.Marshal(buildEntry)
	if err != nil {
		return err
	}

	logrus.Infof("Pulling docker image from %s...", buildImage)
	_, err = d.execDockerCommand("pull", buildImage)
	if err != nil {
		return fmt.Errorf("failed to pull user image %v", err)
	}

	dockerCommandArgs := []string{"container", "run"}
	dockerCommandOptions := []string{"--rm", "-v", srcVol, "-v", artVol, "-v", binVol, "-v", habVol, buildImage}
	configJSONArg := string(configJSON)
	if d.interactMode {
		configJSONArg = fmt.Sprintf("'%s'", configJSONArg)
	}
	launchCommands := []string{"/opt/sd/local_run.sh", configJSONArg, buildEntry.JobName, environment["SD_API_URL"], environment["SD_STORE_URL"], logfilePath}
	if d.interactMode {
		dockerCommandOptions = append([]string{"-itd"}, dockerCommandOptions...)
		dockerCommandOptions = append(dockerCommandOptions, "/bin/sh")
	} else {
		dockerCommandOptions = append(dockerCommandOptions, launchCommands...)
	}

	if buildEntry.MemoryLimit != "" {
		dockerCommandOptions = append([]string{fmt.Sprintf("-m%s", buildEntry.MemoryLimit)}, dockerCommandOptions...)
	}

	if buildEntry.UsePrivileged {
		dockerCommandOptions = append([]string{"--privileged"}, dockerCommandOptions...)
	}

	if d.interactMode {
		// attach build container for sd-local interact mode
		cid, err := d.execDockerCommand(append(dockerCommandArgs, dockerCommandOptions...)...)
		if err != nil {
			return fmt.Errorf("failed to run build container: %v", err)
		}

		attachCommands := []string{"attach", cid}
		commands := [][]string{
			launchCommands,
			{"set", "-a"},
			{".", "/tmp/sd-local.env"},
			{"set", "+a"},
			{"export", "PS1=#"},
			{"cd", "$SD_CHECKOUT_DIR"},
		}
		err = d.attachDockerCommand(attachCommands, commands)
		if err != nil {
			return fmt.Errorf("failed to attach build container: %v", err)
		}
	} else {
		// run for sd-local build mode
		_, err = d.execDockerCommand(append(dockerCommandArgs, dockerCommandOptions...)...)
		if err != nil {
			return fmt.Errorf("failed to run build container: %v", err)
		}
	}

	return nil
}

func (d *docker) attachDockerCommand(attachCommands []string, commands [][]string) error {
	attachCommands = append([]string{"docker"}, attachCommands...)
	if d.useSudo {
		attachCommands = append([]string{"sudo"}, attachCommands...)
	}
	c := execCommand(attachCommands[0], attachCommands[1:]...)

	if d.flagVerbose {
		logrus.Infof("$ %s", c.String())
	}

	return d.interact.Run(c, commands)
}

func (d *docker) execDockerCommand(args ...string) (string, error) {
	commands := append([]string{"docker"}, args...)
	if d.useSudo {
		commands = append([]string{"sudo"}, commands...)
	}
	cmd := execCommand(commands[0], commands[1:]...)
	if d.flagVerbose {
		logrus.Infof("$ %s", strings.Join(commands, " "))
	}
	cmd.Stderr = logrus.StandardLogger().WriterLevel(logrus.ErrorLevel)
	d.commands = append(d.commands, cmd)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	out, err := cmd.Output()
	if d.flagVerbose {
		logrus.Infof("%s", out)
	}
	if err != nil {
		io.Copy(os.Stderr, buf)
		return strings.TrimRight(string(out), "\n"), err
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func (d *docker) kill(sig os.Signal) {
	killedCmds := make([]*exec.Cmd, 0, 10)

	for _, v := range d.commands {
		var err error
		d.mutex.Lock()
		if v.ProcessState != nil {
			continue
		}
		d.mutex.Unlock()

		if d.useSudo {
			cmd := execCommand("sudo", "kill", fmt.Sprintf("-%v", signum(sig)), strconv.Itoa(v.Process.Pid))
			err = cmd.Run()
		} else {
			err = v.Process.Signal(sig)
		}

		if err != nil {
			logrus.Warn(fmt.Errorf("failed to stop process: %v", err))
		} else {
			killedCmds = append(killedCmds, v)
		}
	}

	err := d.waitForProcess(killedCmds)
	if err != nil {
		logrus.Warn(err)
	}
}

func (d *docker) clean() {
	_, err := d.execDockerCommand("volume", "rm", "--force", d.volume)

	if err != nil {
		logrus.Warn(fmt.Errorf("failed to remove volume: %v", err))
	}

	_, err = d.execDockerCommand("volume", "rm", "--force", d.habVolume)

	if err != nil {
		logrus.Warn(fmt.Errorf("failed to remove hab volume: %v", err))
	}
}

func (d *docker) waitForProcess(cmds []*exec.Cmd) error {
	// Reducing this value will make the test faster.
	// However, be sure to specify a time when you can sufficiently confirm that the process is dead.
	t := time.NewTicker(1 * time.Second)
	const retryMax = 9
	retryCnt := 0
	for {
		select {
		case <-t.C:

			retryCnt++
			finish := true

			for _, v := range cmds {
				d.mutex.Lock()
				if v.ProcessState == nil {
					finish = false
				}
				d.mutex.Unlock()
			}
			if finish {
				return nil
			}

			if retryCnt > retryMax {
				return fmt.Errorf("waited %d seconds and could not confirm that the process was dead", retryMax+1)
			}
		}
	}
}

func signum(sig os.Signal) int {
	const numSig = 65

	switch sig := sig.(type) {
	case syscall.Signal:
		i := int(sig)
		if i < 0 || i >= numSig {
			return -1
		}
		return i
	default:
		return -1
	}
}
