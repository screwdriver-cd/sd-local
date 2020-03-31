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
        "github.com/otiai10/copy"
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

type git struct {
	baseDir   string
	remoteURL string
	branch    string
	localPath string
}

type file struct {
	baseDir   string
	srcPath   string
	localPath string
}

func createLocalPath(baseDir string) (string, error) {
	localPath := filepath.Join(baseDir, "repo", strconv.Itoa(rand.Int()))
	err := osMkdirAll(localPath, 0777)
	if err != nil {
		return "", fmt.Errorf("failed to make local source directory: %w", err)
	}

	return localPath, nil
}

// New create new SCM instance
func NewGit(baseDir, srcURL string) (SCM, error) {
	results := srcURLRegex.FindStringSubmatch(srcURL)

	if len(results) == 0 {
		return nil, fmt.Errorf("failed to fetch source code with invalid URL: %s", srcURL)
	}

	remoteURL, branch := results[1], results[2]

	localPath, err := createLocalPath(baseDir)

	if err != nil {
		return nil, err
	}

	g := &git{
		baseDir:   baseDir,
		remoteURL: remoteURL,
		branch:    branch,
		localPath: localPath,
	}

	return g, nil
}

func NewFile(baseDir, srcPath string) (SCM, error) {
	localPath, err := createLocalPath(baseDir)

	if err != nil {
		return nil, err
	}

	f := &file{
		baseDir:   baseDir,
		srcPath:   srcPath,
		localPath: localPath,
	}

	return f, nil
}

func (g *git) Pull() error {
	args := []string{"clone"}
	if g.branch != "" {
		args = append(args, "-b", g.branch)
	}
	args = append(args, g.remoteURL, g.LocalPath())

	cmd := execCommand("git", args...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone remote repository: %w", err)
	}

	return nil
}

func (g *git) Clean() error {
	err := os.RemoveAll(g.LocalPath())
	if err != nil {
		return fmt.Errorf("failed to remove local source directory: %w", err)
	}
	return nil
}

func (g *git) LocalPath() string {
	return g.localPath
}

func (f *file) Pull() error {
        if err := copy.Copy(f.srcPath, f.LocalPath()); err != nil {
		return fmt.Errorf("failed to copy the source directory: %w", err)
	}

	return nil
}

func (f *file) Clean() error {
	err := os.RemoveAll(f.LocalPath())
	if err != nil {
		return fmt.Errorf("failed to remove local source directory: %w", err)
	}
	return nil
}

func (f *file) LocalPath() string {
	return f.localPath
}
