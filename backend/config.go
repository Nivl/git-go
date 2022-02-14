package backend

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/spf13/afero"
)

func (b *Backend) loadConfig() error {
	return nil
}

// InitOptions represents all the options that can be used to
// create a repository
type InitOptions struct {
	// CreateSymlink will create a .git FILE that will contains a path
	// to the repo.
	CreateSymlink bool
}

// Init initializes a repository.
// This method cannot be called concurrently with other methods.
// Calling this method on an existing repository is safe. It will not
// overwrite things that are already there, but will add what's missing.
func (b *Backend) Init(branchName string) error {
	return b.InitWithOptions(branchName, InitOptions{})
}

// InitWithOptions initializes a repository using the provided options.
//
// This method cannot be called concurrently with other methods.
// Calling this method on an existing repository is safe. It will not
// overwrite things that are already there, but will add what's missing.
func (b *Backend) InitWithOptions(branchName string, opts InitOptions) error {
	if opts.CreateSymlink {
		linkSource := filepath.Join(b.config.WorkTreePath, config.DefaultDotGitDirName)
		linkTarget := fmt.Sprintf("gitdir: %s", ginternals.DotGitPath(b.config))
		err := afero.WriteFile(b.fs, linkSource, []byte(linkTarget), 0o644)
		if err != nil {
			return fmt.Errorf("could not create symlink %s: %w", linkSource, err)
		}
	}

	// Create the directories if they don't already exist
	dirs := []string{
		b.Path(),
		ginternals.TagsPath(b.config),
		ginternals.DotGitPath(b.config),
		ginternals.LocalBranchesPath(b.config),
		ginternals.ObjectsPath(b.config),
		ginternals.ObjectsInfoPath(b.config),
		ginternals.ObjectsPacksPath(b.config),
	}
	for _, d := range dirs {
		if err := b.fs.MkdirAll(d, 0o750); err != nil {
			return fmt.Errorf("could not create directory %s: %w", d, err)
		}
	}

	// Create the files with the default content if they don't already exist
	// (taken from a repo created on github)
	files := []struct {
		path    string
		content []byte
	}{
		{
			path:    ginternals.DescriptionFilePath(b.config),
			content: []byte("Unnamed repository; edit this file 'description' to name the repository.\n"),
		},
	}
	for _, f := range files {
		err := afero.WriteFile(b.fs, f.path, f.content, 0o644)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("could not create file %s: %w", f.path, err)
		}
	}

	// We only create a config file if we don't already have one
	_, err := b.fs.Stat(b.config.LocalConfig)
	if errors.Is(err, os.ErrNotExist) {
		if err = b.config.FromFile().Save(); err != nil {
			return fmt.Errorf("could not save the config: %w", err)
		}
	}

	// Create HEAD if it doesn't exist yet
	ref := ginternals.NewSymbolicReference(ginternals.Head, ginternals.LocalBranchFullName(branchName))
	err = b.WriteReferenceSafe(ref)
	if err != nil && !errors.Is(err, ginternals.ErrRefExists) {
		return fmt.Errorf("could not write HEAD: %w", err)
	}

	return nil
}
