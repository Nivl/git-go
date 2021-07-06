package pathutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrNoRepo is an error returned when no repo are found
var ErrNoRepo = errors.New("not a git repository (or any of the parent directories)")

// WorkingTree returns the absolute path to the working tree
func WorkingTree(dotGitDirName string) (path string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current working directory: %w", err)
	}
	return WorkingTreeFromPath(wd, dotGitDirName)
}

// WorkingTreeFromPath returns the absolute path to the root of a repo containing
// the provided directory
func WorkingTreeFromPath(p, dotGitDirName string) (path string, err error) {
	prev := ""
	for p != prev {
		info, err := os.Stat(filepath.Join(p, dotGitDirName))
		if err == nil && info.IsDir() {
			return p, nil
		}

		prev = p
		p = filepath.Dir(p)
	}
	return "", ErrNoRepo
}
