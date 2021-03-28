// Package fsbackend contains an implementation of the backend.Backend
// interface for the filesystem
package fsbackend

import (
	"os"
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
	objectMu     *syncutil.NamedMutex
	cache        *cache.LRU
	looseObjects *sync.Map

	packfiles map[ginternals.Oid]*packfile.Pack

	refs *sync.Map

	// we use afero.Fs instead of the regular os package
	// because some part of the library requires an Fs object.
	// It's easier and cleaner to use the same thing everywhere.
	fs             afero.Fs
	root           string
	objectsDirPath string
}

// New returns a new Backend object
func New(dotGitPath string) (*Backend, error) {
	return NewWithObjectsPath(dotGitPath, filepath.Join(dotGitPath, gitpath.ObjectsPath))
}

// NewWithObjectsPath returns a new backend object that stores object at
// the given paths
func NewWithObjectsPath(dotGitPath, dotGitObjectsPath string) (*Backend, error) {
	c, err := cache.NewLRU(1000)
	if err != nil {
		return nil, xerrors.Errorf("could not create LRU cache: %w", err)
	}
	b := &Backend{
		fs:             afero.NewOsFs(),
		root:           dotGitPath,
		objectsDirPath: dotGitObjectsPath,
		cache:          c,
		objectMu:       syncutil.NewNamedMutex(101),
		packfiles:      map[ginternals.Oid]*packfile.Pack{},
		refs:           &sync.Map{},
		looseObjects:   &sync.Map{},
	}

	// we load a few things in memory
	var loadRefsErr error
	var loadPackErr error
	var loadLooseObjectErr error
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		loadRefsErr = b.loadRefs()
	}()
	go func() {
		defer wg.Done()
		loadLooseObjectErr = b.loadLooseObject()
	}()
	go func() {
		defer wg.Done()
		loadPackErr = b.loadPacks()
	}()
	wg.Wait()

	if loadRefsErr != nil {
		return nil, xerrors.Errorf("could not load references: %w", loadRefsErr)
	}
	if loadPackErr != nil {
		return nil, xerrors.Errorf("could not load packs: %w", loadPackErr)
	}
	if loadLooseObjectErr != nil {
		return nil, xerrors.Errorf("could not load loose objects: %w", loadLooseObjectErr)
	}

	return b, nil
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

// Path returns the absolute path of the repo
func (b *Backend) Path() string {
	return b.root
}

// ObjectsPath returns the absolute path of the object directory
func (b *Backend) ObjectsPath() string {
	return b.objectsDirPath
}

// Init initializes a repository
// This method cannot be called concurrently with other methods
func (b *Backend) Init() error {
	// Create the directories
	dirs := []string{
		filepath.Join(b.root, gitpath.RefsTagsPath),
		filepath.Join(b.root, gitpath.RefsHeadsPath),
		b.objectsDirPath,
		filepath.Join(b.objectsDirPath, gitpath.ObjectsInfoPath),
		filepath.Join(b.objectsDirPath, gitpath.ObjectsPackPath),
	}
	for _, d := range dirs {
		if err := b.fs.MkdirAll(d, 0o750); err != nil {
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
		if err := os.WriteFile(fullPath, f.content, 0o644); err != nil {
			return xerrors.Errorf("could not create file %s: %w", f, err)
		}
	}

	err := b.setDefaultCfg()
	if err != nil {
		return xerrors.Errorf("could not set the default config: %w", err)
	}

	ref := ginternals.NewSymbolicReference(ginternals.Head, gitpath.LocalBranch(ginternals.Master))
	if err := b.WriteReferenceSafe(ref); err != nil {
		return xerrors.Errorf("could not write HEAD: %w", err)
	}

	return nil
}
