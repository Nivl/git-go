package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	dir, cleanup := testutil.TempDir(t)
	t.Cleanup(cleanup)

	// the common dir can be set by creating a file with a path in it
	// we create a `with_commondir` directory to be able to run test in this
	// context
	gitDirWithCommonDir := filepath.Join(dir, "with_commondir")
	err := os.Mkdir(gitDirWithCommonDir, 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(gitDirWithCommonDir, "commondir"), []byte(filepath.Join(dir, "common")), 0o644)
	require.NoError(t, err)

	// git allows the .git to be a special "git-symlink" file that contains
	// a path to the actual repo. see `git init --separate-git-dir`
	// https://git-scm.com/docs/git-init#Documentation/git-init.txt---separate-git-dirltgitdirgt
	wtWithGitfile := filepath.Join(dir, "with_gitfile_symlink")
	err = os.Mkdir(wtWithGitfile, 0o755)
	require.NoError(t, err)
	content := "gitdir: " + filepath.Join(dir, ".git")
	err = os.WriteFile(filepath.Join(wtWithGitfile, DefaultDotGitDirName), []byte(content), 0o644)
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	currentRepoRoot := filepath.Join(cwd, "..", "..")

	validRepoRoot, cleanup := testutil.UnTar(t, testutil.RepoSmall)
	t.Cleanup(cleanup)

	testCases := []struct {
		desc           string
		cfg            LoadConfigOptions
		e              *env.Env
		expectedParams *Config
		expectedError  error
	}{
		{
			desc: "everything default (current repo must be checked out)",
			cfg:  LoadConfigOptions{},
			e:    env.NewFromKVList([]string{}),
			expectedParams: &Config{
				WorkTreePath:     currentRepoRoot,
				GitDirPath:       filepath.Join(currentRepoRoot, DefaultDotGitDirName),
				CommonDirPath:    filepath.Join(currentRepoRoot, DefaultDotGitDirName),
				LocalConfig:      filepath.Join(currentRepoRoot, DefaultDotGitDirName, defaultConfigDirName),
				ObjectDirPath:    filepath.Join(currentRepoRoot, DefaultDotGitDirName, defaultObjectsDirName),
				Prefix:           "",
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
		{
			desc:           "Should fail specifying a work tree (env) without a git path",
			cfg:            LoadConfigOptions{},
			e:              env.NewFromKVList([]string{"GIT_WORK_TREE=" + cwd}),
			expectedParams: &Config{},
			expectedError:  ErrNoWorkTreeAlone,
		},
		{
			desc: "Should fail specifying a work tree (override) without a git path",
			cfg: LoadConfigOptions{
				WorkTreePath: cwd,
			},
			e:              env.NewFromKVList([]string{}),
			expectedParams: &Config{},
			expectedError:  ErrNoWorkTreeAlone,
		},
		{
			desc: "Env should be used when available",
			cfg:  LoadConfigOptions{},
			e: env.NewFromKVList([]string{
				"GIT_WORK_TREE=" + filepath.Join(dir, "wt"),
				"GIT_DIR=" + filepath.Join(dir, "git"),
				"GIT_OBJECT_DIRECTORY=" + filepath.Join(dir, "objects"),
				"GIT_CONFIG=" + filepath.Join(dir, "gitconfig"),
				"PREFIX=" + filepath.Join(dir, "sysconf"),
				"GIT_CONFIG_NOSYSTEM=1",
			}),
			expectedParams: &Config{
				WorkTreePath:     filepath.Join(dir, "wt"),
				GitDirPath:       filepath.Join(dir, "git"),
				CommonDirPath:    filepath.Join(dir, "git"),
				LocalConfig:      filepath.Join(dir, "gitconfig"),
				ObjectDirPath:    filepath.Join(dir, "objects"),
				Prefix:           filepath.Join(dir, "sysconf"),
				SkipSystemConfig: true,
			},
			expectedError: nil,
		},
		{
			desc: "options should override everything",
			cfg: LoadConfigOptions{
				WorkTreePath: filepath.Join(dir, "custom", "wt"),
				GitDirPath:   filepath.Join(dir, "custom", "git"),
			},
			e: env.NewFromKVList([]string{
				"GIT_WORK_TREE=" + filepath.Join(dir, "wt"),
				"GIT_DIR=" + filepath.Join(dir, "git"),
				"GIT_OBJECT_DIRECTORY=" + filepath.Join(dir, "objects"),
				"GIT_CONFIG=" + filepath.Join(dir, "gitconfig"),
				"PREFIX=" + filepath.Join(dir, "sysconf"),
			}),
			expectedParams: &Config{
				WorkTreePath:     filepath.Join(dir, "custom", "wt"),
				GitDirPath:       filepath.Join(dir, "custom", "git"),
				CommonDirPath:    filepath.Join(dir, "custom", "git"),
				LocalConfig:      filepath.Join(dir, "gitconfig"),
				ObjectDirPath:    filepath.Join(dir, "objects"),
				Prefix:           filepath.Join(dir, "sysconf"),
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
		{
			desc: "Should work overriding the working directory",
			cfg: LoadConfigOptions{
				WorkingDirectory: validRepoRoot,
			},
			e: env.NewFromKVList([]string{}),
			expectedParams: &Config{
				WorkTreePath:     validRepoRoot,
				GitDirPath:       filepath.Join(validRepoRoot, DefaultDotGitDirName),
				CommonDirPath:    filepath.Join(validRepoRoot, DefaultDotGitDirName),
				LocalConfig:      filepath.Join(validRepoRoot, DefaultDotGitDirName, defaultConfigDirName),
				ObjectDirPath:    filepath.Join(validRepoRoot, DefaultDotGitDirName, defaultObjectsDirName),
				Prefix:           "",
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
		{
			desc: "relative paths should be made absolute based on the current working directory",
			cfg:  LoadConfigOptions{},
			e: env.NewFromKVList([]string{
				"GIT_WORK_TREE=wt",
				"GIT_DIR=git",
				"GIT_OBJECT_DIRECTORY=objects",
				"GIT_CONFIG=gitconfig",
			}),
			expectedParams: &Config{
				WorkTreePath:  filepath.Join(cwd, "wt"),
				GitDirPath:    filepath.Join(cwd, "git"),
				CommonDirPath: filepath.Join(cwd, "git"),
				LocalConfig:   filepath.Join(cwd, "gitconfig"),
				ObjectDirPath: filepath.Join(cwd, "objects"),
			},
			expectedError: nil,
		},
		{
			desc: "relative working directory should be made absolute based on the working directory",
			cfg: LoadConfigOptions{
				WorkingDirectory: "wd",
			},
			e: env.NewFromKVList([]string{
				"GIT_WORK_TREE=wt",
				"GIT_DIR=git",
				"GIT_OBJECT_DIRECTORY=objects",
				"GIT_CONFIG=gitconfig",
			}),
			expectedParams: &Config{
				WorkTreePath:  filepath.Join(cwd, "wd", "wt"),
				GitDirPath:    filepath.Join(cwd, "wd", "git"),
				CommonDirPath: filepath.Join(cwd, "wd", "git"),
				LocalConfig:   filepath.Join(cwd, "wd", "gitconfig"),
				ObjectDirPath: filepath.Join(cwd, "wd", "objects"),
			},
			expectedError: nil,
		},
		{
			desc: "Custom common dir",
			cfg: LoadConfigOptions{
				WorkTreePath: dir,
				GitDirPath:   filepath.Join(dir, DefaultDotGitDirName),
			},
			e: env.NewFromKVList([]string{
				"GIT_COMMON_DIR=" + filepath.Join(dir, "common"),
			}),
			expectedParams: &Config{
				WorkTreePath:     dir,
				GitDirPath:       filepath.Join(dir, DefaultDotGitDirName),
				CommonDirPath:    filepath.Join(dir, "common"),
				LocalConfig:      filepath.Join(dir, "common", defaultConfigDirName),
				ObjectDirPath:    filepath.Join(dir, "common", defaultObjectsDirName),
				Prefix:           "",
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
		{
			desc: "common dir from file",
			cfg: LoadConfigOptions{
				WorkTreePath: dir,
				GitDirPath:   filepath.Join(dir, "with_commondir"),
			},
			e: env.NewFromKVList([]string{}),
			expectedParams: &Config{
				WorkTreePath:     dir,
				GitDirPath:       gitDirWithCommonDir,
				CommonDirPath:    filepath.Join(dir, "common"),
				LocalConfig:      filepath.Join(dir, "common", defaultConfigDirName),
				ObjectDirPath:    filepath.Join(dir, "common", defaultObjectsDirName),
				Prefix:           "",
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
		{
			desc: "Custom relative common dir",
			cfg: LoadConfigOptions{
				WorkTreePath: dir,
				GitDirPath:   filepath.Join(dir, DefaultDotGitDirName),
			},
			e: env.NewFromKVList([]string{
				"GIT_COMMON_DIR=" + "common",
			}),
			expectedParams: &Config{
				WorkTreePath:     dir,
				GitDirPath:       filepath.Join(dir, DefaultDotGitDirName),
				CommonDirPath:    filepath.Join(dir, DefaultDotGitDirName, "common"),
				LocalConfig:      filepath.Join(dir, DefaultDotGitDirName, "common", defaultConfigDirName),
				ObjectDirPath:    filepath.Join(dir, DefaultDotGitDirName, "common", defaultObjectsDirName),
				Prefix:           "",
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
		{
			desc: "bare repo should not point to a .git",
			cfg: LoadConfigOptions{
				WorkingDirectory: "wd",
				IsBare:           true,
			},
			e: env.NewFromKVList([]string{}),
			expectedParams: &Config{
				WorkTreePath:  "",
				GitDirPath:    filepath.Join(cwd, "wd"),
				CommonDirPath: filepath.Join(cwd, "wd"),
				LocalConfig:   filepath.Join(cwd, "wd", "config"),
				ObjectDirPath: filepath.Join(cwd, "wd", "objects"),
			},
			expectedError: nil,
		},
		{
			desc: "repo with a .git file should work",
			cfg: LoadConfigOptions{
				WorkingDirectory: wtWithGitfile,
			},
			e: env.NewFromKVList([]string{}),
			expectedParams: &Config{
				WorkTreePath:  wtWithGitfile,
				GitDirPath:    filepath.Join(dir, ".git"),
				CommonDirPath: filepath.Join(dir, ".git"),
				LocalConfig:   filepath.Join(dir, ".git", "config"),
				ObjectDirPath: filepath.Join(dir, ".git", "objects"),
			},
			expectedError: nil,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			out, err := LoadConfig(tc.e, tc.cfg)
			if tc.expectedError != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// We don't want to check for files or FS
			out.fromFiles = nil
			out.FS = nil
			out.env = nil

			assert.Equal(t, tc.expectedParams, out)
		})
	}
}

func TestNewGitOptionsSkipEnv(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	currentRepoRoot := filepath.Join(cwd, "..", "..")

	testCases := []struct {
		desc           string
		cfg            LoadConfigOptions
		expectedParams *Config
		expectedError  error
	}{
		{
			desc: "everything default (current repo must be checked out)",
			cfg:  LoadConfigOptions{},
			expectedParams: &Config{
				WorkTreePath:     currentRepoRoot,
				GitDirPath:       filepath.Join(currentRepoRoot, DefaultDotGitDirName),
				CommonDirPath:    filepath.Join(currentRepoRoot, DefaultDotGitDirName),
				LocalConfig:      filepath.Join(currentRepoRoot, DefaultDotGitDirName, defaultConfigDirName),
				ObjectDirPath:    filepath.Join(currentRepoRoot, DefaultDotGitDirName, defaultObjectsDirName),
				Prefix:           "",
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			out, err := LoadConfigSkipEnv(tc.cfg)
			if tc.expectedError != nil {
				require.Error(t, err)
				return
			}

			// We remove some data to make the assertion easier
			out.FS = nil
			out.fromFiles = nil
			out.env = nil

			require.NoError(t, err)
			assert.Equal(t, tc.expectedParams, out)
		})
	}
}

// This test isn't ran in parallel because it share a state between all
// its subtests
func TestWrapper(t *testing.T) {
	// First we setup a config file and we make sure it's loaded
	dir, err := os.Getwd()
	require.NoError(t, err)
	root := filepath.VolumeName(dir) + string(os.PathSeparator)

	expectedWorktreePath := filepath.Join(root, "some", "path")

	// create the config file
	f, cleanup := testutil.TempFile(t)
	t.Cleanup(cleanup)
	_, err = f.WriteString("[core]\nworktree = " + expectedWorktreePath)
	require.NoError(t, err)
	err = f.Sync()
	require.NoError(t, err)

	e := env.NewFromKVList([]string{
		"GIT_CONFIG=" + f.Name(),
	})
	opts := LoadConfigOptions{
		GitDirPath: filepath.Join(root, DefaultDotGitDirName),
	}
	out, err := LoadConfig(e, opts)

	require.NoError(t, err)
	require.Equal(t, expectedWorktreePath, out.WorkTreePath)

	// All good, we can run our other tests that requires a config

	t.Run("FromFile", func(t *testing.T) {
		require.NotNil(t, out.FromFile(), "expected FromFile() to return an object")
	})

	t.Run("Save and Reload", func(t *testing.T) {
		// Validate the original settings
		v, found := out.FromFile().IsBare()
		require.True(t, found, "IsBare should have been set by LoadConfig()")
		require.False(t, v, "IsBare should be false")

		// Updating should update the config in memory
		// a reload should erase the settings
		out.FromFile().UpdateIsBare(true)
		v, found = out.FromFile().IsBare()
		assert.True(t, found, "IsBare should now be set")
		assert.True(t, v, "IsBare should be true")
		require.NoError(t, out.Reload())
		_, found = out.FromFile().IsBare()
		require.False(t, found, "IsBare shouldn't be set after a reload")

		// A reload should keep the value is it's been saved
		out.FromFile().UpdateIsBare(true)
		require.NoError(t, out.fromFiles.Save())
		require.NoError(t, out.Reload())
		v, found = out.FromFile().IsBare()
		assert.True(t, found, "IsBare should now be set")
		assert.True(t, v, "IsBare should be true")
	})
}
