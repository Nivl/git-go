package env

import (
	"path/filepath"
	"strings"

	"github.com/Nivl/git-go/internal/gitpath"
)

// GitOptions represents the options that can be set using the env
type GitOptions struct {
	// GitDirPath represents the path to the .git directory
	// Defaults to .git
	// Maps to GIT_DIR
	GitDirPath string
	// GitObjectDirPath represents the path to the .git/objects directory
	// Defaults to .git/objects
	// Maps to GIT_OBJECT_DIRECTORY
	GitObjectDirPath string
	// GitConfig represents the config file to load
	// Defaults to .git/config
	// Maps to GIT_CONFIG
	GitConfig string
	// SkipSystemConfig states whether we should use the system config or not
	// Defaults to false
	// Maps to GIT_CONFIG_NOSYSTEM
	SkipSystemConfig bool

	// if set it means the exported values are final
	isFinalized bool
}

// NewGitOptions returns a new GitOptions that fetches the data from the
// env
//
// Usage: NewGitOptions(NewFromOs())
func NewGitOptions(e *Env) *GitOptions {
	SkipSystemConfig := false
	switch strings.ToLower(e.Get("GIT_CONFIG_NOSYSTEM")) {
	case "yes", "1", "true":
		SkipSystemConfig = true
	}

	return &GitOptions{
		GitDirPath:       e.Get("GIT_DIR"),
		GitObjectDirPath: e.Get("GIT_OBJECT_DIRECTORY"),
		SkipSystemConfig: SkipSystemConfig,
		GitConfig:        e.Get("GIT_CONFIG"),
	}
}

// FinalizeOptions represents all the options available to finalize
// the GitOptions
type FinalizeOptions struct {
	ProjectPath string
	IsBare      bool
}

// Finalize gathers all the data and finalize the GitOptions object
// so all the paths are set with their locations, etc.
func (opts *GitOptions) Finalize(p FinalizeOptions) {
	if opts.isFinalized {
		return
	}

	opts.isFinalized = true
	opts.GitDirPath = opts.buildDotGitPath(p.ProjectPath, p.IsBare)
	opts.GitObjectDirPath = opts.buildDotGitObjectsPath(p.ProjectPath, opts.GitDirPath)
	if opts.GitConfig == "" {
		opts.GitConfig = filepath.Join(opts.GitDirPath, gitpath.ConfigPath)
	}
}

// IsFinalized returns whether the options have been finalized
func (opts *GitOptions) IsFinalized() bool {
	if opts == nil {
		return false
	}
	return opts.isFinalized
}

// buildDotGitPath returns the absolute path to the .git directory
// In its most basic configuration, projectPath is the folder containing
// the .git directory
func (opts *GitOptions) buildDotGitPath(projectPath string, isBare bool) string {
	dotGitPath := projectPath
	if !isBare {
		dotGitPath = filepath.Join(projectPath, gitpath.DotGitPath)
	}
	// if GitDirPath is set then it doesn't matter if the repo is bare
	// or not. It actually doesn't make sense to set this repo as bare if
	// we're going to provide a GitDirPath.
	if opts.GitDirPath != "" {
		dotGitPath = opts.GitDirPath
		if !filepath.IsAbs(opts.GitDirPath) {
			dotGitPath = filepath.Join(projectPath, opts.GitDirPath)
		}
	}
	return dotGitPath
}

// buildDotGitObjectsPath returns the absolute path to the .git/objects directory
// In its most basic configuration, projectPath is the folder containing
// the .git directory
func (opts *GitOptions) buildDotGitObjectsPath(projectPath, dotGitPath string) string {
	gitObjectsPath := filepath.Join(dotGitPath, gitpath.ObjectsPath)
	if opts.GitObjectDirPath != "" {
		gitObjectsPath = opts.GitObjectDirPath
		if !filepath.IsAbs(opts.GitObjectDirPath) {
			gitObjectsPath = filepath.Join(projectPath, opts.GitObjectDirPath)
		}
	}
	return gitObjectsPath
}
