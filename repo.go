package git

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
)

// List of errors returned by the Repository struct
var (
	ErrRepositoryNotExist           = errors.New("repository does not exist")
	ErrRepositoryUnsupportedVersion = errors.New("repository nor supported")
)

// Repository represent a git repository
// A Git repository is the .git/ folder inside a project.
// This repository tracks all changes made to files in your project,
// building a history over time.
// https://blog.axosoft.com/learning-git-repository/
type Repository struct {
	path        string
	projectPath string
}

// NewRepository creates a new Repository instance. The repository
// will neither be created nor loaded. You must call Load() or Init()
// to get a full repository
func NewRepository(projectPath string) *Repository {
	return &Repository{
		path:        filepath.Join(projectPath, DotGitPath),
		projectPath: projectPath,
	}
}

// LoadRepository loads an existing git repository by reading its
// config file, and returns a Repository instance
func LoadRepository(projectPath string) (*Repository, error) {
	r := NewRepository(projectPath)
	return r, r.Load()
}

// InitRepository initialize a new git repository by creating the .git
// directory in the given path, which is where almost everything that
// Git stores and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
func InitRepository(projectPath string) (*Repository, error) {
	r := NewRepository(projectPath)
	return r, r.Init()
}

// Init initialize a new git repository by creating the .git directory
// in the given path, which is where almost everything that Git stores
// and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
func (r *Repository) Init() error {
	// We don't check if the repo already exist, and just add what's
	// missing to prevent corruption (ie. if we fail creating the repo
	// retrying would always fail because the .git already exist,
	// and Load() would always fail because the repo is invalid)

	// Create all the default folders
	dirs := []string{
		BranchesPath,
		ObjectsPath,
		RefsTagsPath,
		RefsHeadsPath,
		ObjectsInfoPath,
		ObjectsPackPath,
	}
	for _, d := range dirs {
		fullPath := filepath.Join(r.path, d)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return errors.Wrapf(err, "could not create directory %s", d)
		}
	}

	// Create the files with the default content
	// (taken from a repo created on github)
	files := []struct {
		path    string
		content []byte
	}{
		{
			path:    DescriptionPath,
			content: []byte("Unnamed repository; edit this file 'description' to name the repository.\n"),
		},
		{
			path:    HEADPath,
			content: []byte("ref: refs/heads/master\n"),
		},
	}

	for _, f := range files {
		fullPath := filepath.Join(r.path, f.path)
		if err := ioutil.WriteFile(fullPath, f.content, 0644); err != nil {
			return errors.Wrapf(err, "could not create file %s", f)
		}
	}

	if err := r.setDefaultCfg(); err != nil {
		return errors.Wrap(err, "could not create config file")
	}

	return nil
}

// Load loads an existing git repository by reading its config file,
// and returns a Repository instance
func (r *Repository) Load() error {
	// we first make sure the repo exist
	_, err := os.Stat(r.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "could not check for %s directory", DotGitPath)
		}
		return ErrRepositoryNotExist
	}

	// Load the config file
	// https://git-scm.com/docs/git-config
	cfg, err := ini.Load(filepath.Join(r.path, ConfigPath))
	if err != nil {
		return errors.Wrapf(err, "could not read config file")
	}

	// Validate the config
	repoVersion := cfg.Section(cfgCore).Key(cfgCoreFormatVersion).MustInt(0)
	if repoVersion != 0 {
		return ErrRepositoryUnsupportedVersion
	}

	return nil
}

// setDefaultCfg set and persists the default git configuration for
// the repository
func (r *Repository) setDefaultCfg() error {
	cfg := ini.Empty()

	// Core
	core, err := cfg.NewSection(cfgCore)
	if err != nil {
		return errors.Wrap(err, "could not create core section")
	}
	coreCfg := map[string]string{
		cfgCoreFormatVersion:     "0",
		cfgCoreFileMode:          "true",
		cfgCoreBare:              "false",
		cfgCoreLogAllRefUpdate:   "true",
		cfgCoreIgnoreCase:        "true",
		cfgCorePrecomposeUnicode: "true",
	}
	for k, v := range coreCfg {
		if _, err := core.NewKey(k, v); err != nil {
			return errors.Wrapf(err, "could not set %s", k)
		}
	}
	return cfg.SaveTo(filepath.Join(r.path, ConfigPath))
}
