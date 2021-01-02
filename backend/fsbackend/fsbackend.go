// Package fsbackend contains an implementation of the backend.Backend
// interface for the filesystem
package fsbackend

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/packfile"
	"github.com/Nivl/git-go/internal/cache"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/syncutil"
	"github.com/spf13/afero"
	"golang.org/x/xerrors"
)

// we make sure the struct implements the interface
var _ backend.Backend = (*Backend)(nil)

// Backend is a Backend implementation that uses the filesystem to store data
type Backend struct {
	// we use afero.Fs instead of the regular os package
	// because some part of the library requires an Fs object.
	// It's easier and cleaner to use the same thing everywhere.
	fs   afero.Fs
	root string

	objectMu *syncutil.NamedMutex
	cache    *cache.LRU

	packfileParsing sync.Once
	packfiles       map[ginternals.Oid]*packfile.Pack

	refMu *syncutil.NamedMutex
}

// New returns a new Backend object
func New(dotGitPath string) *Backend {
	return &Backend{
		fs:        afero.NewOsFs(),
		root:      dotGitPath,
		cache:     cache.NewLRU(1000),
		objectMu:  syncutil.NewNamedMutex(101),
		refMu:     syncutil.NewNamedMutex(101),
		packfiles: map[ginternals.Oid]*packfile.Pack{},
	}
}

// Close frees the resources used by the Backend
// This method cannot be called concurrently with other methods
func (b *Backend) Close() (err error) {
	for oid, pack := range b.packfiles {
		if e := pack.Close(); e != nil {
			// we don't return directly because we still want to try to
			// close the other packfiles
			err = xerrors.Errorf("could not close packfile %s: %w", oid.String(), err)
		}
	}
	b.packfiles = map[ginternals.Oid]*packfile.Pack{}
	return err
}

// Init initializes a repository
// This method cannot be called concurrently with other methods
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
		if err := b.fs.MkdirAll(fullPath, 0o750); err != nil {
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
