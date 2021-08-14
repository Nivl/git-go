package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Nivl/git-go/env"
	"gopkg.in/ini.v1"
)

// defaultLoadOption contains the params used to load the config files
//nolint:gochecknoglobals // It's a global because we
// don't want to have to redefine it all the time.
// Treat this as a const, don't ever change it from a method, even for
// testing.
var defaultLoadOption = ini.LoadOptions{
	SkipUnrecognizableLines: true,
}

// FileAggregate represents the aggregate of all the config files
// impacting a repository
type FileAggregate struct {
	cfg *Config
	agg *ini.File
}

// RepoFormatVersion returns the version of the format of the repo
func (cfg *FileAggregate) RepoFormatVersion() (version int, ok bool) {
	v, err := cfg.agg.Section("core").Key("repositoryformatversion").Int()
	if err != nil {
		return 0, false
	}
	return v, true
}

// DefaultBranch returns the branch name to use when creating a new
// repository.
// The branch name isn't checked and may be an invalid value
func (cfg *FileAggregate) DefaultBranch() (name string, ok bool) {
	v := cfg.agg.Section("init").Key("defaultBranch").String()
	if v == "" {
		return "", false
	}
	return v, true
}

// WorkTree returns the path of the work-tree
func (cfg *FileAggregate) WorkTree() (workTree string, ok bool) {
	v := cfg.agg.Section("core").Key("worktree").String()
	return v, v != ""
}

// NewFileAggregate loads all the available config files and returns an object
// with accessor
func NewFileAggregate(e *env.Env, cfg *Config) (confFile *FileAggregate, err error) {
	confFile = &FileAggregate{
		cfg: cfg,
	}
	configPaths := getPaths(e, cfg)

	// Because we want to use afero instead of the file system, we cannot
	// just provide the the file paths to ini.Load. Instead we need to open
	// all the files ourselves, provide the files to ini, and close everything.
	// We use []interface{} because "ini.Load" wants a slice of interfaces
	files := make([]interface{}, 0, len(configPaths))
	for _, p := range configPaths {
		_, sErr := cfg.FS.Stat(p)
		if sErr != nil {
			// not every config files are expected to exists on disk
			// so we skip all the one that doesn't
			if errors.Is(sErr, os.ErrNotExist) {
				continue
			}
			err = fmt.Errorf("could not check file %s: %w", p, sErr)
			break
		}

		f, fErr := cfg.FS.Open(p)
		if fErr != nil {
			err = fmt.Errorf("could not open file %s: %w", p, fErr)
			break
		}
		files = append(files, f)
	}
	defer func() {
		// we need to cleanup the file descriptors to avoid a leak
		for _, f := range files {
			//nolint:errcheck // it's expected to fail as the files are already closed.
			// go-ini already closes the files for us. This code is
			// only here to prevent a FD leak in case go-ini updates the
			// behavior and we don't see it / remember about it
			f.(io.ReadCloser).Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		confFile.agg = ini.Empty(defaultLoadOption)
		return confFile, nil
	}

	// ini.Load wants the config file separated over 2 args, the second args
	// being a spreadable.
	src := files[0]
	others := files[1:]
	confFile.agg, err = ini.LoadSources(defaultLoadOption, src, others...)
	if err != nil {
		return nil, fmt.Errorf("could not load config file: %w", err)
	}
	return confFile, nil
}

func appendIfValid(array *[]string, envVar string, p ...string) {
	if envVar != "" {
		*array = append(*array, filepath.Join(envVar, filepath.Join(p...)))
	}
}

func getPaths(e *env.Env, cfg *Config) []string {
	configPaths := []string{}

	// system
	// git looks for a file located ar $(prefix)/etc/gitconfig, which is
	// a value provided at compile time or through the env ($PREFIX).
	// Since we often don't have this value set, we'll do a
	// system-specific brute-force later on if $PREFIX isn't set.
	if !cfg.SkipSystemConfig && cfg.Prefix != "" {
		configPaths = append(configPaths, filepath.Join(cfg.Prefix, "etc", "gitconfig"))
	}

	switch runtime.GOOS {
	case "windows":
		// system
		if !cfg.SkipSystemConfig && cfg.Prefix == "" {
			appendIfValid(&configPaths, e.Get("ALLUSERSPROFILE"), "Application Data", "Git", "config")
			appendIfValid(&configPaths, e.Get("ProgramFiles(x86)"), "Git", "etc", "gitconfig")
			appendIfValid(&configPaths, e.Get("ProgramFiles"), "Git", "mingw64", "etc", "gitconfig")
		}
		// global
		appendIfValid(&configPaths, e.Get("USERPROFILE"), ".gitconfig")
	default:
		// System
		if !cfg.SkipSystemConfig && cfg.Prefix == "" {
			configPaths = append(configPaths,
				"/etc/gitconfig",
				"/usr/local/etc/gitconfig",
				"/opt/homebrew/etc/gitconfig",
			)
		}
		// global
		if e.Get("XDG_CONFIG_HOME") != "" {
			configPaths = append(configPaths, filepath.Join(e.Get("XDG_CONFIG_HOME"), "git", ".gitconfig"))
		} else {
			appendIfValid(&configPaths, e.Get("HOME"), ".config", ".git", ".gitconfig")
		}
	}
	// shared global
	appendIfValid(&configPaths, e.Get("HOME"), ".gitconfig")
	// local
	configPaths = append(configPaths, cfg.LocalConfig)
	return configPaths
}
