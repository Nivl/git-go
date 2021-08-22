package backend

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/spf13/afero"
	"gopkg.in/ini.v1"
)

func (b *Backend) loadConfig() error {
	return nil
}

// Init initializes a repository.
// This method cannot be called concurrently with other methods.
// Calling this method on an existing repository is safe. It will not
// overwrite things that are already there, but will add what's missing.
func (b *Backend) Init(branchName string) error {
	return b.InitWithSymlink(branchName, false)
}

// InitWithSymlink initializes a repository. If createSymlink is set
// to true, a text file will be created that will point to the repo.
//
// This method cannot be called concurrently with other methods.
// Calling this method on an existing repository is safe. It will not
// overwrite things that are already there, but will add what's missing.
func (b *Backend) InitWithSymlink(branchName string, createSymlink bool) error {
	if createSymlink {
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
		if err = b.setDefaultCfg(); err != nil {
			return fmt.Errorf("could not set the default config: %w", err)
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

// setDefaultCfg set and persists the default git configuration for
// the repository
func (b *Backend) setDefaultCfg() error {
	cfg := ini.Empty()

	// Core
	core, err := cfg.NewSection(CfgCore)
	if err != nil {
		return fmt.Errorf("could not create core section: %w", err)
	}
	coreCfg := map[string]string{
		CfgCoreFormatVersion:     "0",
		CfgCoreFileMode:          "true",
		CfgCoreBare:              "false",
		CfgCoreLogAllRefUpdate:   "true",
		CfgCoreIgnoreCase:        "true",
		CfgCorePrecomposeUnicode: "true",
	}
	for k, v := range coreCfg {
		if _, e := core.NewKey(k, v); e != nil {
			return fmt.Errorf("could not set %s: %w", k, e)
		}
	}
	// TODO(melvin): this should use the provided FS interface instead
	// of using actual file system
	err = cfg.SaveTo(b.config.LocalConfig)
	if err != nil {
		return fmt.Errorf("could not persist file: %w", err)
	}
	return nil
}
