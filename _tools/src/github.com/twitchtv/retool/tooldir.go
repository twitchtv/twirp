package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

const (
	toolDirName = "_tools"
)

var (
	baseDir = flag.String("base-dir", "",
		"Path of project root.  If not specified and the working directory is within a git repository, the root of "+
			"the repository is used.  If the working directory is not within a git repository, the working directory "+
			"is used.")
	toolDir = flag.String("tool-dir", "",
		"Path where tools are stored.  The default value is the subdirectory of -base-dir named '_tools'.")

	// These globals are set by ensureTooldir() after factoring in the flags above.
	baseDirPath string
	toolDirPath string
)

// If the working directory is within a git repository, return the path of the repository's root; otherwise, return the
// empty string.  An error is returned iff invoking 'git' fails for some other reason.
func getRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	stdout, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitStatus := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
			if exitStatus == 128 { // not in a repository
				return "", nil
			}
		}
		return "", errors.Wrap(err, "failed to invoke git")
	}
	repoRoot := strings.TrimSpace(string(stdout))
	return repoRoot, nil
}

func ensureTooldir() error {
	var err error

	baseDirPath = *baseDir
	if baseDirPath == "" {
		var repoRootPath string
		repoRootPath, err = getRepoRoot()
		if err != nil {
			return errors.Wrap(err, "failed to check for enclosing git repository")
		}
		if repoRootPath == "" {
			baseDirPath, err = os.Getwd()
			if err != nil {
				return errors.Wrap(err, "failed to get working directory")
			}
		} else {
			baseDirPath = repoRootPath
		}
	}

	toolDirPath = *toolDir
	if toolDirPath == "" {
		toolDirPath = filepath.Join(baseDirPath, toolDirName)
	}

	verbosef("base dir: %v\n", baseDirPath)
	verbosef("tool dir: %v\n", toolDirPath)

	stat, err := os.Stat(toolDirPath)
	switch {
	case os.IsNotExist(err):
		err = os.Mkdir(toolDirPath, 0777)
		if err != nil {
			return errors.Wrap(err, "unable to create tooldir")
		}
	case err != nil:
		return errors.Wrap(err, "unable to stat tool directory")
	case !stat.IsDir():
		return errors.New("tool directory already exists, but it is not a directory; you can use -tool-dir to change where tools are saved")
	}

	err = ioutil.WriteFile(path.Join(toolDirPath, ".gitignore"), gitignore, 0664)
	if err != nil {
		return errors.Wrap(err, "unable to update .gitignore")
	}

	return nil
}

var gitignore = []byte(strings.TrimLeft(`
/bin/
/pkg/
/manifest.json
`, "\n"))
