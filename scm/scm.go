package scm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

var (
	srcUrlRegex = regexp.MustCompile(`^((?:(?:https://(?:[^@/:\s]+@)?)|git@)+([^/:\s]+)(?:/|:)([^/:\s]+)/([^\s]+?)(?:\.git))?(#[^\s]+)?$`)
	osMkdirAll  = os.MkdirAll
	execCommand = exec.Command
)

type SCM interface {
	Pull() error
	Clean() error
	LocalPath() string
}

type scm struct {
	baseDir   string
	remoteUrl string
	instance  string
	org       string
	repo      string
	branch    string
}

func New(baseDir, srcUrl string) (SCM, error) {
	results := srcUrlRegex.FindStringSubmatch(srcUrl)
	remoteUrl, instance, org, repo, branch := results[1], results[2], results[3], results[4], results[5]

	scm := &scm{
		baseDir:   baseDir,
		remoteUrl: remoteUrl,
		instance:  instance,
		org:       org,
		repo:      repo,
	}
	if branch != "" {
		scm.branch = branch[1:]
	}

	fmt.Println(scm.LocalPath())
	err := osMkdirAll(scm.LocalPath(), 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to make local source directory: %w", err)
	}

	return scm, nil
}

func (scm *scm) Pull() error {
	args := []string{"clone"}
	if scm.branch != "" {
		args = append(args, "-b", scm.branch)
	}
	args = append(args, scm.remoteUrl)
	fmt.Println(args)
	cmd := execCommand("git", args...)
	cmd.Dir = filepath.Join(scm.baseDir, scm.instance, scm.org)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone remote repository: %w", err)
	}

	return nil
}

func (scm *scm) Clean() error {
	err := os.RemoveAll(scm.LocalPath())
	if err != nil {
		return fmt.Errorf("failed to remove local source directory: %w", err)
	}
	return nil
}

func (scm *scm) LocalPath() string {
	return filepath.Join(scm.baseDir, "repo", scm.instance, scm.org, scm.repo)
}
