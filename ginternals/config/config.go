// Package config contains structs to interact with git configuration
// as well as to configure the library
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/pathutil"
	"github.com/spf13/afero"
)

// ErrNoWorkTreeAlone is thrown when a work tree path is given without
// a git path
var ErrNoWorkTreeAlone = errors.New("cannot specify a work tree without also specifying a git dir")

// Config represents the config of a repository, whether it's from
// the various config files or from the options that can be set using
// the env https://git-scm.com/book/en/v2/Git-Internals-Environment-Variables
//
// If you decide to create a Config by yourself, make sure to set correct
// values everywhere
type Config struct {
	// FS represents the file system implementation to use to look for
	// files and directories.
	// Defaults to the regular filesystem.
	FS afero.Fs

	// fromFiles contains a reference to the config values held in
	// files
	fromFiles *FileAggregate

	// GitDirPath represents the path to the .git directory
	// Maps to $GIT_DIR if set
	// Defaults to finding a ".git" folder in the current directory,
	// going up in the tree until reaching /
	GitDirPath string
	// WorkTreePath represents the path to the .git directory
	// Maps to $GIT_WORK_TREE
	// Defaults to $(GitDirPath)/.. or $(current-dir) depending on if
	// GitDirPath was set or not.
	WorkTreePath string
	// ObjectDirPath represents the path to the .git/objects directory
	// Maps to $GIT_OBJECT_DIRECTORY
	// Defaults to $(GitDirPath)/.git/objects
	ObjectDirPath string
	// LocalConfig represents the config file to load
	// Maps to $GIT_CONFIG
	// Defaults to $(GitDirPath)/config if not sets
	LocalConfig string
	// Prefix contains the base for finding the system configuration file.
	// $(prefix)/etc/gitconfig
	// Maps to $PREFIX
	// Defaults to an empty string
	Prefix string
	// SkipSystemConfig states whether we should use the system config or not
	// Maps to $GIT_CONFIG_NOSYSTEM
	// Defaults to false
	SkipSystemConfig bool
}

// LoadConfigOptions represents all the params used to set the default
// values of a Config object
type LoadConfigOptions struct {
	// FS represents the file system implementation to use to look for
	// files and directories.
	// Defaults to the regular filesystem.
	FS afero.Fs
	// WorkingDirectory represents the current working directory
	// Defaults to the current working directory
	WorkingDirectory string
	// WorkTreePath corresponds to the directory that should contain the .git.
	// Set this value to change the default behavior and overwrite
	// $GIT_WORK_TREE.
	WorkTreePath string
	// GitDirPath corresponds to the .git directory
	// Set this value to change the default behavior and overwrite
	// $GIT_DIR.
	GitDirPath string
	// IsBare defines if the repo is base. It means that the repo ha no
	// work tree
	IsBare bool
	// SkipGitDirLookUp will disable automatic lookup of the .git directory.
	// Defaults to false which means that if no path is provided
	// to $GitDirPath or $GIT_DIR, the method will look for a .git dir in
	// $WorkingDirectory and will go up the tree until it finds one.
	//
	// You should only set this value to true if you want to initialize a
	// new repository.
	SkipGitDirLookUp bool
}

// LoadConfig returns a new Config that fetches the data from the
// env
// This is what you want to use to give your users some control over
// git.
// If you want something more direct without control, use NewGitOptionsSkipEnv()
func LoadConfig(e *env.Env, p LoadConfigOptions) (*Config, error) {
	SkipSystemConfig := false
	switch strings.ToLower(e.Get("GIT_CONFIG_NOSYSTEM")) {
	case "yes", "1", "true":
		SkipSystemConfig = true
	}

	opts := &Config{
		GitDirPath:       e.Get("GIT_DIR"),
		WorkTreePath:     e.Get("GIT_WORK_TREE"),
		ObjectDirPath:    e.Get("GIT_OBJECT_DIRECTORY"),
		SkipSystemConfig: SkipSystemConfig,
		LocalConfig:      e.Get("GIT_CONFIG"),
		Prefix:           e.Get("PREFIX"),
	}

	if err := setConfig(e, opts, p); err != nil {
		return nil, err
	}
	return opts, nil
}

// LoadConfigSkipEnv returns a new Config that skips the env
// and uses the default values
func LoadConfigSkipEnv(opts LoadConfigOptions) (*Config, error) {
	return LoadConfig(env.NewFromKVList([]string{}), opts)
}

func setConfig(e *env.Env, p *Config, opts LoadConfigOptions) (err error) {
	if opts.FS == nil {
		opts.FS = afero.NewOsFs()
	}
	p.FS = opts.FS

	// FIXME(melvin): Ultimately we should get this from afero, but
	// there are no methods for that
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get the current directory: %w", err)
	}
	if opts.WorkingDirectory == "" {
		opts.WorkingDirectory = wd
	}
	if !filepath.IsAbs(opts.WorkingDirectory) {
		opts.WorkingDirectory = filepath.Join(wd, opts.WorkingDirectory)
	}

	// $GIT_WORK_TREE and --work-tree cannot be set if $GIT_DIR or
	// --git-dir isn't set. core.worktree isn't affected
	if opts.GitDirPath == "" && p.GitDirPath == "" && (opts.WorkTreePath != "" || p.WorkTreePath != "") {
		return ErrNoWorkTreeAlone
	}

	// GirDir rules:
	// - p.GitDirPath contains either nothing or $GIT_DIR
	// - opts.GitDirPath contains either nothing or a value used to override
	//   p.GitDirPath.
	// - If nothing set, a .git directory will looked for by walking up the
	//   current directory.
	// - If relative, the path will be appended to the current working
	//   directory.
	if opts.GitDirPath != "" {
		p.GitDirPath = opts.GitDirPath
	}
	guessedWorkingTree := opts.WorkingDirectory
	switch p.GitDirPath {
	default:
		if !filepath.IsAbs(p.GitDirPath) {
			p.GitDirPath = filepath.Join(opts.WorkingDirectory, p.GitDirPath)
		}
	case "":
		if !opts.SkipGitDirLookUp {
			guessedWorkingTree, err = pathutil.WorkingTreeFromPath(opts.WorkingDirectory)
			if err != nil {
				return fmt.Errorf("could not find working tree: %w", err)
			}
		}
		p.GitDirPath = filepath.Join(guessedWorkingTree, gitpath.DotGitPath)
	}

	// LocalConfig rules:
	// - p.LocalConfig contains either nothing or a path to the .git/config
	// - Fallback to $(GitDirPath)/config
	//
	// If relative, the path will be appended to the current working
	// directory.
	if p.LocalConfig == "" {
		p.LocalConfig = filepath.Join(p.GitDirPath, gitpath.ConfigPath)
	}
	if !filepath.IsAbs(p.LocalConfig) {
		p.LocalConfig = filepath.Join(opts.WorkingDirectory, p.LocalConfig)
	}

	// ObjectDirPath rules:
	// - p.ObjectDirPath contains either nothing or a path to the .git/objects
	// - Fallback to $(GitDirPath)/objects
	//
	// If relative, the path will be appended to the current working
	// directory.
	if p.ObjectDirPath == "" {
		p.ObjectDirPath = filepath.Join(p.GitDirPath, gitpath.ObjectsPath)
	}
	if !filepath.IsAbs(p.ObjectDirPath) {
		p.ObjectDirPath = filepath.Join(opts.WorkingDirectory, p.ObjectDirPath)
	}

	p.fromFiles, err = NewFileAggregate(e, p)
	if err != nil {
		return fmt.Errorf("could not load config files: %w", err)
	}

	// Worktree rules:
	//
	// - core.Worktree contains either nothing or the default path to
	// the working tree.
	// - p.WorkTreePath contains either nothing, $GIT_WORK_TREE.
	//	 It overrides core.Worktree
	// - opts.WorkTreePath contains either nothing or a path to the
	//   working tree.
	//   It overrides p.WorkTreePath
	// - guessedWorkingTree contains either nothing or the path containing
	//	 the .git directory.
	//   It's use as fallback for opts.WorkTreePath
	// - Fallback on the current working directory
	//
	// If any path are relative, they will be relative to the current
	// working directory
	if path, ok := p.fromFiles.WorkTree(); ok {
		p.WorkTreePath = path
	}
	if opts.WorkTreePath != "" {
		p.WorkTreePath = opts.WorkTreePath
	}
	// if the repo is bare then we don't automatically set a working tree
	// if none are provided
	if p.WorkTreePath == "" && !opts.IsBare {
		p.WorkTreePath = guessedWorkingTree
	}
	if p.WorkTreePath != "" && !filepath.IsAbs(p.WorkTreePath) {
		p.WorkTreePath = filepath.Join(opts.WorkingDirectory, p.WorkTreePath)
	}

	return nil
}
