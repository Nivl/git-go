package backend

import (
	"os"
	"path/filepath"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/gitpath"
	"golang.org/x/xerrors"
	"gopkg.in/ini.v1"
)

func (b *Backend) loadConfig() error {
	return nil
}

// Init initializes a repository
// This method cannot be called concurrently with other methods
func (b *Backend) Init() error {
	// Create the directories
	dirs := []string{
		b.Path(),
		filepath.Join(b.Path(), gitpath.RefsTagsPath),
		filepath.Join(b.Path(), gitpath.RefsHeadsPath),
		b.ObjectsPath(),
		filepath.Join(b.ObjectsPath(), gitpath.ObjectsInfoPath),
		filepath.Join(b.ObjectsPath(), gitpath.ObjectsPackPath),
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
		fullPath := filepath.Join(b.Path(), f.path)
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

// setDefaultCfg set and persists the default git configuration for
// the repository
func (b *Backend) setDefaultCfg() error {
	cfg := ini.Empty()

	// Core
	core, err := cfg.NewSection(CfgCore)
	if err != nil {
		return xerrors.Errorf("could not create core section: %w", err)
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
		if _, err := core.NewKey(k, v); err != nil {
			return xerrors.Errorf("could not set %s: %w", k, err)
		}
	}
	return cfg.SaveTo(filepath.Join(b.Path(), gitpath.ConfigPath))
}
