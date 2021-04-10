package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitParams(t *testing.T) {
	t.Parallel()

	// To be able to build an absolute path on Windows we need to know
	// the Volume name
	dir, err := os.Getwd()
	require.NoError(t, err)
	root := filepath.VolumeName(dir) + string(os.PathSeparator)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	currentRepoRoot := filepath.Join(cwd, "..", "..")

	validRepoRoot, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)

	testCases := []struct {
		desc           string
		cfg            config.NewGitParamsOptions
		e              *env.Env
		expectedParams *config.GitParams
		expectedError  error
	}{
		{
			desc: "everything default (current repo must be checked out)",
			cfg:  config.NewGitParamsOptions{},
			e:    env.NewFromKVList([]string{}),
			expectedParams: &config.GitParams{
				WorkTreePath:     currentRepoRoot,
				GitDirPath:       filepath.Join(currentRepoRoot, gitpath.DotGitPath),
				LocalConfig:      filepath.Join(currentRepoRoot, gitpath.DotGitPath, gitpath.ConfigPath),
				ObjectDirPath:    filepath.Join(currentRepoRoot, gitpath.DotGitPath, gitpath.ObjectsPath),
				Prefix:           "",
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
		{
			desc:           "Should fail specifying a work tree (env) without a git path",
			cfg:            config.NewGitParamsOptions{},
			e:              env.NewFromKVList([]string{"GIT_WORK_TREE=" + cwd}),
			expectedParams: &config.GitParams{},
			expectedError:  config.ErrNoWorkTreeAlone,
		},
		{
			desc: "Should fail specifying a work tree (override) without a git path",
			cfg: config.NewGitParamsOptions{
				WorkTreePath: cwd,
			},
			e:              env.NewFromKVList([]string{}),
			expectedParams: &config.GitParams{},
			expectedError:  config.ErrNoWorkTreeAlone,
		},
		{
			desc: "Env should be used when available",
			cfg:  config.NewGitParamsOptions{},
			e: env.NewFromKVList([]string{
				"GIT_WORK_TREE=" + filepath.Join(root, "wt"),
				"GIT_DIR=" + filepath.Join(root, "git"),
				"GIT_OBJECT_DIRECTORY=" + filepath.Join(root, "objects"),
				"GIT_CONFIG=" + filepath.Join(root, "gitconfig"),
				"PREFIX=" + filepath.Join(root, "sysconf"),
				"GIT_CONFIG_NOSYSTEM=1",
			}),
			expectedParams: &config.GitParams{
				WorkTreePath:     filepath.Join(root, "wt"),
				GitDirPath:       filepath.Join(root, "git"),
				LocalConfig:      filepath.Join(root, "gitconfig"),
				ObjectDirPath:    filepath.Join(root, "objects"),
				Prefix:           filepath.Join(root, "sysconf"),
				SkipSystemConfig: true,
			},
			expectedError: nil,
		},
		{
			desc: "options should override everything",
			cfg: config.NewGitParamsOptions{
				WorkTreePath: filepath.Join(root, "custom", "wt"),
				GitDirPath:   filepath.Join(root, "custom", "git"),
			},
			e: env.NewFromKVList([]string{
				"GIT_WORK_TREE=" + filepath.Join(root, "wt"),
				"GIT_DIR=" + filepath.Join(root, "git"),
				"GIT_OBJECT_DIRECTORY=" + filepath.Join(root, "objects"),
				"GIT_CONFIG=" + filepath.Join(root, "gitconfig"),
				"PREFIX=" + filepath.Join(root, "sysconf"),
			}),
			expectedParams: &config.GitParams{
				WorkTreePath:     filepath.Join(root, "custom", "wt"),
				GitDirPath:       filepath.Join(root, "custom", "git"),
				LocalConfig:      filepath.Join(root, "gitconfig"),
				ObjectDirPath:    filepath.Join(root, "objects"),
				Prefix:           filepath.Join(root, "sysconf"),
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
		{
			desc: "Should work overriding the working directory",
			cfg: config.NewGitParamsOptions{
				WorkingDirectory: validRepoRoot,
			},
			e: env.NewFromKVList([]string{}),
			expectedParams: &config.GitParams{
				WorkTreePath:     filepath.Join(validRepoRoot),
				GitDirPath:       filepath.Join(validRepoRoot, gitpath.DotGitPath),
				LocalConfig:      filepath.Join(validRepoRoot, gitpath.DotGitPath, gitpath.ConfigPath),
				ObjectDirPath:    filepath.Join(validRepoRoot, gitpath.DotGitPath, gitpath.ObjectsPath),
				Prefix:           "",
				SkipSystemConfig: false,
			},
			expectedError: nil,
		},
		{
			desc: "relative paths should be made absolute based on the current working directory",
			cfg:  config.NewGitParamsOptions{},
			e: env.NewFromKVList([]string{
				"GIT_WORK_TREE=wt",
				"GIT_DIR=git",
				"GIT_OBJECT_DIRECTORY=objects",
				"GIT_CONFIG=gitconfig",
			}),
			expectedParams: &config.GitParams{
				WorkTreePath:  filepath.Join(cwd, "wt"),
				GitDirPath:    filepath.Join(cwd, "git"),
				LocalConfig:   filepath.Join(cwd, "gitconfig"),
				ObjectDirPath: filepath.Join(cwd, "objects"),
			},
			expectedError: nil,
		},
		{
			desc: "relative working directory should be made absolute based on the working directory",
			cfg: config.NewGitParamsOptions{
				WorkingDirectory: "wd",
			},
			e: env.NewFromKVList([]string{
				"GIT_WORK_TREE=wt",
				"GIT_DIR=git",
				"GIT_OBJECT_DIRECTORY=objects",
				"GIT_CONFIG=gitconfig",
			}),
			expectedParams: &config.GitParams{
				WorkTreePath:  filepath.Join(cwd, "wd", "wt"),
				GitDirPath:    filepath.Join(cwd, "wd", "git"),
				LocalConfig:   filepath.Join(cwd, "wd", "gitconfig"),
				ObjectDirPath: filepath.Join(cwd, "wd", "objects"),
			},
			expectedError: nil,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			out, err := config.NewGitParams(tc.e, tc.cfg)
			if tc.expectedError != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
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
		cfg            config.NewGitParamsOptions
		expectedParams *config.GitParams
		expectedError  error
	}{
		{
			desc: "everything default (current repo must be checked out)",
			cfg:  config.NewGitParamsOptions{},
			expectedParams: &config.GitParams{
				WorkTreePath:     currentRepoRoot,
				GitDirPath:       filepath.Join(currentRepoRoot, gitpath.DotGitPath),
				LocalConfig:      filepath.Join(currentRepoRoot, gitpath.DotGitPath, gitpath.ConfigPath),
				ObjectDirPath:    filepath.Join(currentRepoRoot, gitpath.DotGitPath, gitpath.ObjectsPath),
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

			out, err := config.NewGitOptionsSkipEnv(tc.cfg)
			if tc.expectedError != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedParams, out)
		})
	}
}
