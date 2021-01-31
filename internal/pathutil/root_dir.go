package pathutil

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"
)

// ErrNoRepo is an error returned when no repo are found
var ErrNoRepo = errors.New("not a git repository (or any of the parent directories)")

// RepoRoot returns the absolute path to the root of the repo
func RepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", xerrors.Errorf("could not get current working directory: %w", err)
	}
	return RepoRootFromPath(wd)
}

// RepoRootFromPath returns the absolute path to the root of a repo containing
// the provided directory
func RepoRootFromPath(p string) (string, error) {
	prev := ""
	for p != prev {
		// Regular repo
		info, err := os.Stat(filepath.Join(p, ".git"))
		if err == nil && info.IsDir() {
			return p, nil
		}
		// Bare repo
		info, err = os.Stat(filepath.Join(p, "HEAD"))
		if err == nil && !info.IsDir() && info.Size() > 0 {
			return p, nil
		}

		prev = p
		p = filepath.Dir(p)
	}
	return "", ErrNoRepo
}
