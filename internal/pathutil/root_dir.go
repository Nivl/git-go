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

	for wd != string(os.PathSeparator) {
		_, err := os.Stat(filepath.Join(wd, ".git"))
		if err == nil {
			return wd, nil
		}
		wd = filepath.Dir(wd)
	}

	return "", ErrNoRepo
}
