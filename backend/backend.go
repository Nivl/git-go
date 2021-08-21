// Package backend contains structs and methods to store and
// retrieve data from the .git directory
package backend

import (
	"fmt"
	"sync"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/Nivl/git-go/ginternals/packfile"
	"github.com/Nivl/git-go/internal/cache"
	"github.com/Nivl/git-go/internal/syncutil"
	"github.com/spf13/afero"
)

// Backend is a Backend implementation that uses the filesystem to store data
type Backend struct {
	config *config.Config

	objectMu     *syncutil.NamedMutex
	cache        *cache.LRU
	looseObjects *sync.Map

	packfiles map[ginternals.Oid]*packfile.Pack

	refs *sync.Map

	fs afero.Fs

	hashAlgorithm string
}

// NewFS returns a new Backend object using the local FileSystem
func NewFS(cfg *config.Config) (*Backend, error) {
	return New(cfg, afero.NewOsFs())
}

// New returns a new Backend object
func New(cfg *config.Config, fs afero.Fs) (*Backend, error) {
	c, err := cache.NewLRU(1000)
	if err != nil {
		return nil, fmt.Errorf("could not create LRU cache: %w", err)
	}

	b := &Backend{
		config:       cfg,
		fs:           fs,
		cache:        c,
		objectMu:     syncutil.NewNamedMutex(101),
		packfiles:    map[ginternals.Oid]*packfile.Pack{},
		refs:         &sync.Map{},
		looseObjects: &sync.Map{},
	}

	b.hashAlgorithm, _ = cfg.FromFile().Objectformat()
	if b.hashAlgorithm == "" {
		b.hashAlgorithm = "sha1"
	}
	if b.hashAlgorithm != "sha1" && b.hashAlgorithm != "sha256" {
		return nil, ErrUnknownHashAlgo
	}

	// we load a few things in memory
	wg := sync.WaitGroup{}

	var loadRefsErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		loadRefsErr = b.loadRefs()
	}()
	var loadLooseObjectErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		loadLooseObjectErr = b.loadLooseObject()
	}()
	var loadPackErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		loadPackErr = b.loadPacks()
	}()
	var loadConfigErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		loadConfigErr = b.loadConfig()
	}()

	wg.Wait()

	if loadRefsErr != nil {
		return nil, fmt.Errorf("could not load references: %w", loadRefsErr)
	}
	if loadPackErr != nil {
		return nil, fmt.Errorf("could not load packs: %w", loadPackErr)
	}
	if loadLooseObjectErr != nil {
		return nil, fmt.Errorf("could not load loose objects: %w", loadLooseObjectErr)
	}
	if loadConfigErr != nil {
		return nil, fmt.Errorf("could not load config: %w", loadConfigErr)
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
			err = fmt.Errorf("could not close packfile %s: %w", oid.String(), err)
		}
	}
	b.packfiles = map[ginternals.Oid]*packfile.Pack{}
	return err
}

// Path returns the absolute path of the repo
func (b *Backend) Path() string {
	return ginternals.DotGitPath(b.config)
}
