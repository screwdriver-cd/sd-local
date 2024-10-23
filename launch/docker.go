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

// DinD has the information needed to start the dind-rootless container
type DinD struct {
	enabled         bool
	volume          string
	shareVolumeName string
	shareVolumePath string
	container       string
	network         string
	image           string
}

type docker struct {
	volume            string
	habVolume         string
	setupImage        string
	setupImageVersion string
	useSudo           bool
	interactiveMode   bool
	commands          []*exec.Cmd
	mutex             *sync.Mutex
	flagVerbose       bool
	interact          Interacter
	socketPath        string
	localVolumes      []string
	buildUser         string
	noImagePull       bool
	dind              DinD
}

var _ runner = (*docker)(nil)
var execCommand = exec.Command

const (
	// ArtifactsDir is default artifact directory name
	ArtifactsDir = "sd-artifacts"
	// LogFile is default logfile name for build log
	LogFile = "builds.log"
	// StepsDir is default directory name for step script files
	StepsDir = ".sd-steps"
	// The definition of "ScmHost" and "OrgRepo" is in "PipelineFromID" of "screwdriver/screwdriver_local.go"
	scmHost = "screwdriver.cd"
	orgRepo = "sd-local/local-build"
)

func newDocker(setupImage, setupImageVer string, useSudo bool, interactiveMode bool, socketPath string, flagVerbose bool, localVolumes []string, buildUser string, noImagePull bool, dindEnabled bool) runner {
	return &docker{
		volume:            "SD_LAUNCH_BIN",
		habVolume:         "SD_LAUNCH_HAB",
		setupImage:        setupImage,
		setupImageVersion: setupImageVer,
		useSudo:           useSudo,
		interactiveMode:   interactiveMode,
		commands:          make([]*exec.Cmd, 0, 10),
		mutex:             &sync.Mutex{},
		flagVerbose:       flagVerbose,
		interact:          &Interact{flagVerbose: flagVerbose},
		socketPath:        socketPath,
		localVolumes:      localVolumes,
		buildUser:         buildUser,
		noImagePull:       noImagePull,
		dind: DinD{
			enabled:         dindEnabled,
			volume:          "SD_DIND_CERT",
			shareVolumeName: "SD_DIND_SHARE",
			shareVolumePath: "/opt/sd_dind_share",
			container:       "sd-local-dind",
			network:         "sd-local-dind-bridge",
			image:           "docker:23.0.1-dind-rootless",
		},
	}
}

func (d *docker) setupBin() error {
	mount := fmt.Sprintf("%s:/opt/sd/", d.volume)
	habMount := fmt.Sprintf("%s:/hab", d.habVolume)
	image := fmt.Sprintf("%s:%s", d.setupImage, d.setupImageVersion)

	if !d.noImagePull {
		_, err := d.execDockerCommand("pull", image)
		if err != nil {
			return fmt.Errorf("failed to pull launcher image: %v", err)
		}
	}

	// The mechanism for population is that VOLUMEs were declared in the image, so they copy what was in their layer to
	// the mounted location on first mount of non-existing volumes
	// NOTE: docker allows copying to first-time mounted as well, but both docker and podman copy to non-existing ones.
	//       therefore, volumes are not pre-created, but created on first mention by the image that populates them
	//       and then used by subsequent images that then use their content.
	_, err := d.execDockerCommand("container", "run", "--rm", "--pull", "never", "-v", mount, "-v", habMount, "--entrypoint", "/bin/echo", image, "set up bin")
	if err != nil {
		return fmt.Errorf("failed to prepare build scripts: %v", err)
	}

	return nil
}

func setupInteractiveMode(buildEntry *buildEntry) error {
	stepsPath, err := filepath.Abs(StepsDir)

	if err != nil {
		return err
	}

	if err := os.MkdirAll(stepsPath, 0777); err != nil {
		return err
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/bin", stepsPath), 0777); err != nil {
		return err
	}

	shellBin := GetEnv(buildEntry.Environment, "USER_SHELL_BIN")

	if len(shellBin) == 0 {
		shellBin = "/bin/sh"
	}

	sdRunShell := fmt.Sprintf(`#!%s
step_name="$1"
step_list=$(ls ${SD_STEPS_DIR})
if [ "${step_name}" = "" ]; then echo "${step_list}";
else . "${SD_STEPS_DIR}/${step_name}"; fi
`, shellBin)

	if err := os.WriteFile(fmt.Sprintf("%s/bin/sd-run", stepsPath), []byte(sdRunShell), 0755); err != nil {
		return err
	}

	for _, step := range buildEntry.Steps {
		if err := os.WriteFile(fmt.Sprintf("%s/%s", stepsPath, step.Name), []byte("#!"+shellBin+" -e\n"+step.Command), 0755); err != nil {
			return err
		}
	}

	// Overwrite steps for sd-local interact mode. The env will load later.
	buildEntry.Steps = []screwdriver.Step{
		{
			Name:    "sd-local-init",
			Command: "export > /tmp/sd-local.env",
		},
	}

	return nil
}

func (d *docker) runBuild(buildEntry buildEntry) error {
	dockerCommandArgs := []string{"container", "run"}
	dockerCommandOptions := []string{"--rm", "--entrypoint", "/bin/sh", "-e", "SSH_AUTH_SOCK=/tmp/auth.sock"}

	if d.dind.enabled {
		if err := d.runDinD(); err != nil {
			return fmt.Errorf("failed to prepare dind container: %v", err)
		}

		dockerCommandOptions = append(
			[]string{
				"--network", d.dind.network,
				"-e", "DOCKER_TLS_CERTDIR=/certs",
				"-e", "DOCKER_HOST=tcp://docker:2376",
				"-e", "DOCKER_TLS_VERIFY=1",
				"-e", "DOCKER_CERT_PATH=/certs/client",
				"-e", fmt.Sprintf("SD_DIND_SHARE_PATH=%s", d.dind.shareVolumePath),
				"-v", fmt.Sprintf("%s:/certs/client:ro", d.dind.volume),
				"-v", fmt.Sprintf("%s:%s", d.dind.shareVolumeName, d.dind.shareVolumePath),
			},
			dockerCommandOptions...)
	}

	environment := buildEntry.Environment

	srcDir := buildEntry.SrcPath
	hostArtDir := buildEntry.ArtifactsPath
	hostStepsDir := buildEntry.StepsPath
	containerArtDir := GetEnv(environment, "SD_ARTIFACTS_DIR")
	containerStepsDir := GetEnv(environment, "SD_STEPS_DIR")
	buildImage := buildEntry.Image
	logfilePath := filepath.Join(containerArtDir, LogFile)

	srcVol := fmt.Sprintf("%s/:/sd/workspace/src/%s/%s", srcDir, scmHost, orgRepo)
	artVol := fmt.Sprintf("%s/:%s", hostArtDir, containerArtDir)
	stpVol := fmt.Sprintf("%s/:%s", hostStepsDir, containerStepsDir)
	binVol := fmt.Sprintf("%s:%s", d.volume, "/opt/sd")
	habVol := fmt.Sprintf("%s:%s", d.habVolume, "/opt/sd/hab")

	dockerVolumes := append(d.localVolumes, srcVol, artVol, stpVol, binVol, habVol, fmt.Sprintf("%s:/tmp/auth.sock:rw", d.socketPath))
	for _, v := range dockerVolumes {
		dockerCommandOptions = append(dockerCommandOptions, "-v", v)
	}

	if buildEntry.MemoryLimit != "" {
		dockerCommandOptions = append(dockerCommandOptions, fmt.Sprintf("-m%s", buildEntry.MemoryLimit))
	}

	if buildEntry.UsePrivileged {
		dockerCommandOptions = append(dockerCommandOptions, "--privileged")
	}

	if d.interactiveMode {
		if err := setupInteractiveMode(&buildEntry); err != nil {
			return err
		}

		dockerCommandOptions = append(dockerCommandOptions, "-itd")
	}

	if d.buildUser != "" {
		dockerCommandOptions = append(dockerCommandOptions, fmt.Sprintf("-u%s", d.buildUser))
	}

	// Disable automatic image pulling
	// Build image is explicitly pulled when the --no-image-pull option is not used
	dockerCommandOptions = append(dockerCommandOptions, "--pull", "never")

	// Now options are "(docker container run) --rm --entry-point ... --pull never <buildImage>"
	dockerCommandOptions = append(dockerCommandOptions, buildImage)

	configJSON, err := json.Marshal(buildEntry)
	if err != nil {
		return err
	}

	configJSONArg := string(configJSON)
	if d.interactiveMode {
		configJSONArg = fmt.Sprintf("%q", configJSONArg)
	}

	launchCommands := []string{
		"/opt/sd/local_run.sh",
		configJSONArg,
		buildEntry.JobName,
		GetEnv(environment, "SD_API_URL"),
		GetEnv(environment, "SD_STORE_URL"),
		logfilePath,
	}

	// Pull build image explicitly before docker run
	if !d.noImagePull {
		logrus.Infof("Pulling docker image from %s...", buildImage)

		if _, err := d.execDockerCommand("pull", buildImage); err != nil {
			return fmt.Errorf("failed to pull user image %v", err)
		}
	}

	if d.interactiveMode {
		// Create build conatiner without command (e.g. docker run --rm ... image_name)
		cid, err := d.execDockerCommand(append(dockerCommandArgs, dockerCommandOptions...)...)
		if err != nil {
			return fmt.Errorf("failed to run build container: %v", err)
		}

		// attach build container for sd-local interact mode
		attachCommands := []string{"attach", cid}
		commands := [][]string{
			launchCommands,
			{"set", "-a"},
			{".", "/tmp/sd-local.env"},
			{"set", "+a"},
			{"export", "PS1='sd-local# '"},
			{"cd", "$SD_CHECKOUT_DIR"},
			{"sdrun() { . /$SD_STEPS_DIR/bin/sd-run $@; }"},
		}

		if err := d.attachDockerCommand(attachCommands, commands); err != nil {
			return fmt.Errorf("failed to attach build container: %v", err)
		}
	} else {
		// Run build container with launch command (e.g. docker run --rm ... image_name /opt/sd/local_run.sh ...)
		dockerCommandOptions = append(dockerCommandOptions, launchCommands...)

		if _, err := d.execDockerCommand(append(dockerCommandArgs, dockerCommandOptions...)...); err != nil {
			return fmt.Errorf("failed to run build container: %v", err)
		}
	}

	return nil
}

func (d *docker) runDinD() error {
	if !d.noImagePull {
		logrus.Infof("Pulling dind image from %s...", d.dind.image)
		_, err := d.execDockerCommand("pull", d.dind.image)
		if err != nil {
			return fmt.Errorf("failed to pull user image %v", err)
		}
	}

	if _, err := d.execDockerCommand([]string{"network", "create", d.dind.network}...); err != nil {
		return fmt.Errorf("failed to create network: %v", err)
	}

	dockerCommandArgs := []string{"container", "run"}
	dockerCommandOptions := []string{
		"--rm",
		"--privileged",
		"--pull", "never",
		"--name", "sd-local-dind",
		"-d",
		"--network", d.dind.network,
		"--network-alias", "docker",
		"-e", "DOCKER_TLS_CERTDIR=/certs",
		"-v", fmt.Sprintf("%s:/certs/client", d.dind.volume),
		"-v", fmt.Sprintf("%s:/opt/sd_dind_share", d.dind.shareVolumeName),
		d.dind.image,
	}

	if _, err := d.execDockerCommand(append(dockerCommandArgs, dockerCommandOptions...)...); err != nil {
		return fmt.Errorf("failed to run dind container: %v", err)
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
	// Since the habVolume is mounted inside the mountpoint for volume, it must be removed first.
	_, err := d.execDockerCommand("volume", "rm", "--force", d.habVolume)

	if err != nil {
		logrus.Warn(fmt.Errorf("failed to remove volume: %v", err))
	}

	_, err = d.execDockerCommand("volume", "rm", "--force", d.volume)

	if err != nil {
		logrus.Warn(fmt.Errorf("failed to remove hab volume: %v", err))
	}

	if d.interactiveMode {
		if stepsPath, err := filepath.Abs(StepsDir); err == nil {
			err := os.RemoveAll(stepsPath)

			if err != nil {
				logrus.Warn(fmt.Errorf("failed to remove sd-steps directory %s: %v", stepsPath, err))
			}
		} else {
			logrus.Warn(fmt.Errorf("failed to parse sd-steps directory %s: %v", stepsPath, err))
		}
	}

	if d.dind.enabled {
		_, err = d.execDockerCommand("kill", d.dind.container)

		if err != nil {
			logrus.Warn(fmt.Errorf("failed to remove dind container: %v", err))
		}

		_, err = d.execDockerCommand("network", "rm", "--force", d.dind.network)

		if err != nil {
			logrus.Warn(fmt.Errorf("failed to remove dind volume: %v", err))
		}

		_, err = d.execDockerCommand("volume", "rm", "--force", d.dind.volume)

		if err != nil {
			logrus.Warn(fmt.Errorf("failed to remove dind volume: %v", err))
		}

		_, err = d.execDockerCommand("volume", "rm", "--force", d.dind.shareVolumeName)

		if err != nil {
			logrus.Warn(fmt.Errorf("failed to remove dind share volume: %v", err))
		}
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

// GetEnv returns the newest value corresponding to the key
func GetEnv(en []map[string]string, key string) string {
	s := ""
	for _, e := range en {
		if v, ok := e[key]; ok {
			s = v
		}
	}
	return s
}
