package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/pathutil"
	"golang.org/x/xerrors"
)

// GitParams represents the options that can be set using the env
// https://git-scm.com/book/en/v2/Git-Internals-Environment-Variables
//
// If you decide to create a GitParams by yourself, make sure to set correct
// values everywhere
type GitParams struct {
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
	// GitConfig represents the config file to load
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

// NewGitParamsOptions represents all the params used to set the default
// values of GitOptions
type NewGitParamsOptions struct {
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

// NewGitParams returns a new GitParams that fetches the data from the
// env
// This is what you want to use to give your users some control over
// git.
// If you want something more direct without control, use NewGitOptionsSkipEnv()
func NewGitParams(e *env.Env, p NewGitParamsOptions) (*GitParams, error) {
	SkipSystemConfig := false
	switch strings.ToLower(e.Get("GIT_CONFIG_NOSYSTEM")) {
	case "yes", "1", "true":
		SkipSystemConfig = true
	}

	opts := &GitParams{
		GitDirPath:       e.Get("GIT_DIR"),
		WorkTreePath:     e.Get("GIT_WORK_TREE"),
		ObjectDirPath:    e.Get("GIT_OBJECT_DIRECTORY"),
		SkipSystemConfig: SkipSystemConfig,
		LocalConfig:      e.Get("GIT_CONFIG"),
		Prefix:           e.Get("PREFIX"),
	}

	if err := setGitParams(opts, p); err != nil {
		return nil, err
	}
	return opts, nil
}

// NewGitOptionsSkipEnv returns a new GitOptions that skips the env
// and uses the default values
func NewGitOptionsSkipEnv(opts NewGitParamsOptions) (*GitParams, error) {
	p := &GitParams{}
	if err := setGitParams(p, opts); err != nil {
		return nil, err
	}
	return p, nil
}

func setGitParams(p *GitParams, opts NewGitParamsOptions) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return xerrors.Errorf("could not get the current directory: %w", err)
	}
	if opts.WorkingDirectory == "" {
		opts.WorkingDirectory = wd
	}
	if !filepath.IsAbs(opts.WorkingDirectory) {
		opts.WorkingDirectory = filepath.Join(wd, opts.WorkingDirectory)
	}

	// $GIT_WORK_TREE and --work-tree cannot be set if $GIT_DIR or
	// --git-dir isn't set
	if opts.GitDirPath == "" && p.GitDirPath == "" && (opts.WorkTreePath != "" || p.WorkTreePath != "") {
		return xerrors.Errorf("cannot specify a work tree without also specifying a git dir: %w", err)
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
				return xerrors.Errorf("could not find working tree: %w", err)
			}
		}
		p.GitDirPath = filepath.Join(guessedWorkingTree, gitpath.DotGitPath)
	}

	// TODO(melvin): that should be where we load the config files so
	// we can access data such as core.Worktree.

	// Worktree rules:
	//
	// - core.Worktree contains either nothing or the default  path to
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
	//
	// TODO(melvin): add support for core.Worktree and $GIT_COMMON_DIR
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

	return nil
}
