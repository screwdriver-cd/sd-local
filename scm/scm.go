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
	Clean() error
	LocalPath() string
}

type scm struct {
	baseDir   string
	remoteURL string
	branch    string
	localPath string
}

// New create new SCM instance
func New(baseDir, srcURL string) (SCM, error) {
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
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone remote repository: %w", err)
	}

	return nil
}

func (s *scm) Clean() error {
	err := os.RemoveAll(s.LocalPath())
	if err != nil {
		return fmt.Errorf("failed to remove local source directory: %w", err)
	}
	return nil
}

func (s *scm) LocalPath() string {
	return s.localPath
}
