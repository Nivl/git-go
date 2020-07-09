// Package fsbackend contains an implementation of the backend.Backend
// interface for the filesystem
package fsbackend

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/internal/gitpath"
	"golang.org/x/xerrors"
)

// we make sure the struct implements the interface
var _ backend.Backend = (*Backend)(nil)

// Backend is a Backend implementation that uses the filesystem to store data
type Backend struct {
	root string
}

// New returns a new Backend object
func New(dotGitPath string) *Backend {
	return &Backend{
		root: dotGitPath,
	}
}

// Init initializes a repository
func (b *Backend) Init() error {
	// Create the directories
	dirs := []string{
		gitpath.ObjectsPath,
		gitpath.RefsTagsPath,
		gitpath.RefsHeadsPath,
		gitpath.ObjectsInfoPath,
		gitpath.ObjectsPackPath,
	}
	for _, d := range dirs {
		fullPath := filepath.Join(b.root, d)
		if err := os.MkdirAll(fullPath, 0o750); err != nil {
			return xerrors.Errorf("could not create directory %s: %w", d, err)
		}
	}

	// Create the files with the default content
	// (taken from a repo created on github)
	files := []struct {
		path    string
		content []byte
	}{
		{
			path:    gitpath.DescriptionPath,
			content: []byte("Unnamed repository; edit this file 'description' to name the repository.\n"),
		},
	}
	for _, f := range files {
		fullPath := filepath.Join(b.root, f.path)
		if err := ioutil.WriteFile(fullPath, f.content, 0o644); err != nil {
			return xerrors.Errorf("could not create file %s: %w", f, err)
		}
	}

	err := b.setDefaultCfg()
	if err != nil {
		return xerrors.Errorf("could not set the default config: %w", err)
	}

	return nil
}
