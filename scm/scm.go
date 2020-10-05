package scm

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	srcURLRegex = regexp.MustCompile(`^((?:(?:https://(?:[^@/:\s]+@)?)|git@)+(?:[^/:\s]+)(?:/|:)(?:[^/:\s]+)/(?:[^\s]+?)(?:\.git)?)(?:#([^\s]+))?$`)
	osMkdirAll  = os.MkdirAll
	execCommand = exec.Command
)

// SCM is able to fetch source code to build
type SCM interface {
	Pull() error
	Kill(os.Signal)
	Clean()
	LocalPath() string
}

type scm struct {
	baseDir   string
	remoteURL string
	branch    string
	localPath string
	commands  []*exec.Cmd
	sudo      bool
}

// New create new SCM instance
func New(baseDir, srcURL string, sudo bool) (SCM, error) {
	results := srcURLRegex.FindStringSubmatch(srcURL)

	if len(results) == 0 {
		return nil, fmt.Errorf("failed to fetch source code with invalid URL: %s", srcURL)
	}

	remoteURL, branch := results[1], results[2]

	s := &scm{
		baseDir:   baseDir,
		remoteURL: remoteURL,
		branch:    branch,
		localPath: filepath.Join(baseDir, "repo", strconv.Itoa(rand.Int())),
		commands:  make([]*exec.Cmd, 0, 10),
		sudo:      sudo,
	}

	err := osMkdirAll(s.LocalPath(), 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to make local source directory: %w", err)
	}

	return s, nil
}

func (s *scm) Pull() error {
	args := []string{"clone"}
	if s.branch != "" {
		args = append(args, "-b", s.branch)
	}
	args = append(args, s.remoteURL, s.LocalPath())

	cmd := execCommand("git", args...)
	s.commands = append(s.commands, cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone remote repository: %w", err)
	}

	return nil
}

func (s *scm) Kill(sig os.Signal) {
	for _, v := range s.commands {
		if v.ProcessState != nil {
			continue
		}
		err := v.Process.Signal(sig)
		if err != nil {
			logrus.Warn(fmt.Errorf("failed to stop process: %v", err))
		}
	}
}

func (s *scm) Clean() {
	commands := []string{"rm", "-rf", s.LocalPath()}
	if s.sudo {
		commands = append([]string{"sudo"}, commands...)
	}
	err := execCommand(commands[0], commands[1:]...).Run()
	if err != nil {
		logrus.Warn(fmt.Errorf("failed to remove local source directory: %w", err))
	}
}

func (s *scm) LocalPath() string {
	return s.localPath
}
