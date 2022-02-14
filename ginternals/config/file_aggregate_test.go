package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileAggregate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc          string
		env           *env.Env
		cfg           *Config
		expectedOut   *FileAggregate
		expectedError error
	}{
		{
			desc: "should work with no files available",
			env:  env.NewFromKVList([]string{}),
			cfg: &Config{
				SkipSystemConfig: true,
				FS:               afero.NewOsFs(),
			},
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			f, err := NewFileAggregate(tc.env, tc.cfg)
			if tc.expectedError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError, "unexpected error")
				require.Nil(t, f)
			} else {
				require.NoError(t, err)
				require.NotNil(t, f)
			}
		})
	}
}

func TestGetters(t *testing.T) {
	t.Parallel()

	// Setup a few config files, a global one and a local one
	dirPath, cleanup := testhelper.TempDir(t)
	t.Cleanup(cleanup)

	err := os.Mkdir(filepath.Join(dirPath, "etc"), 0o755)
	require.NoError(t, err)

	localConfigPath := filepath.Join(dirPath, "local_config")
	globalConfigPath := filepath.Join(dirPath, "etc", "gitconfig")

	err = os.WriteFile(globalConfigPath, []byte(`
	[core]
		worktree = root_dir
	`), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(localConfigPath, []byte(`
	[core]
		worktree = local_dir
		repositoryformatversion = 0
	[init]
		defaultBranch = main
	`), 0o644)
	require.NoError(t, err)

	// Agg contains the config of both files. The local data should
	// override the global ones
	agg, err := NewFileAggregate(env.NewFromKVList([]string{}),
		&Config{
			LocalConfig: localConfigPath,
			FS:          afero.NewOsFs(),
			Prefix:      dirPath,
		})
	require.NoError(t, err)

	// global only contains the global config
	global, err := NewFileAggregate(env.NewFromKVList([]string{}),
		&Config{
			LocalConfig: globalConfigPath,
			FS:          afero.NewOsFs(),
			Prefix:      dirPath,
		})
	require.NoError(t, err)

	t.Run("WorkTree", func(t *testing.T) {
		t.Parallel()
		wt, ok := agg.WorkTree()
		assert.True(t, ok, "expected to find core.worktree")
		assert.Equal(t, "local_dir", wt)
	})

	t.Run("RepoFormatVersion", func(t *testing.T) {
		t.Parallel()

		t.Run("Default", func(t *testing.T) {
			t.Parallel()
			v, ok := global.RepoFormatVersion()
			assert.False(t, ok, "expected to NOT find core.repositoryformatversion")
			assert.Equal(t, 0, v)
		})

		t.Run("With value", func(t *testing.T) {
			t.Parallel()
			v, ok := agg.RepoFormatVersion()
			assert.True(t, ok, "expected to find core.repositoryformatversion")
			assert.Equal(t, 0, v)
		})
	})

	t.Run("defaultBranch", func(t *testing.T) {
		t.Parallel()

		t.Run("Default", func(t *testing.T) {
			t.Parallel()
			v, ok := global.DefaultBranch()
			assert.False(t, ok, "expected to NOT find init.defaultBranch")
			assert.Equal(t, "", v)
		})

		t.Run("With value", func(t *testing.T) {
			t.Parallel()
			v, ok := agg.DefaultBranch()
			assert.True(t, ok, "expected to find init.defaultBranch")
			assert.Equal(t, "main", v)
		})
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	// Setup a few config files, a global one and a local one
	dirPath, cleanup := testhelper.TempDir(t)
	t.Cleanup(cleanup)

	err := os.Mkdir(filepath.Join(dirPath, "etc"), 0o755)
	require.NoError(t, err)

	localConfigPath := filepath.Join(dirPath, "local_config")
	globalConfigPath := filepath.Join(dirPath, "etc", "gitconfig")

	err = os.WriteFile(globalConfigPath, []byte(`
	[core]
		worktree = root_dir
	`), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(localConfigPath, []byte(`
	[core]
		worktree = local_dir
		repositoryformatversion = 0
		bare = false
	[init]
		defaultBranch = main
	`), 0o644)
	require.NoError(t, err)

	// Agg contains the config of both files. The local data should
	// override the global ones
	agg, err := NewFileAggregate(env.NewFromKVList([]string{}),
		&Config{
			LocalConfig: localConfigPath,
			FS:          afero.NewOsFs(),
			Prefix:      dirPath,
		})
	require.NoError(t, err)

	t.Run("IsBare", func(t *testing.T) {
		t.Parallel()

		// We make sure the default data are as we expect
		v, found := agg.IsBare()
		require.True(t, found, "IsBare should be found")
		require.False(t, v, "IsBare should be false")

		// Update should change the value of the config
		agg.UpdateIsBare(true)
		v, found = agg.IsBare()
		assert.True(t, found, "IsBare should be found")
		assert.True(t, v, "IsBare should be true")
	})
}

func TestGetPaths(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "windows":
		t.Run("windows", func(t *testing.T) {
			t.Parallel()
			testCases := []struct {
				desc        string
				env         *env.Env
				cfg         *Config
				expectedOut []string
			}{
				{
					desc: "No env and skip, should return the local file",
					env:  env.NewFromKVList([]string{}),
					cfg: &Config{
						LocalConfig:      "C:\\local\\config",
						SkipSystemConfig: true,
					},
					expectedOut: []string{"C:\\local\\config"},
				},
				{
					desc: "No env and no skip, should return the local",
					env:  env.NewFromKVList([]string{}),
					cfg: &Config{
						LocalConfig:      "C:\\local\\config",
						SkipSystemConfig: false,
					},
					expectedOut: []string{
						"C:\\local\\config",
					},
				},
				{
					desc: "no skip and env should return correct values",
					env: env.NewFromKVList([]string{
						"ALLUSERSPROFILE=C:\\profiles\\all",
						"ProgramFiles(x86)=C:\\ProgramFiles(x86)",
						"ProgramFiles=C:\\ProgramFiles",
						"USERPROFILE=C:\\profiles\\user",
					}),
					cfg: &Config{
						LocalConfig:      "C:\\local\\config",
						SkipSystemConfig: false,
					},
					expectedOut: []string{
						"C:\\profiles\\all\\Application Data\\Git\\config",
						"C:\\ProgramFiles(x86)\\Git\\etc\\gitconfig",
						"C:\\ProgramFiles\\Git\\mingw64\\etc\\gitconfig",
						"C:\\profiles\\user\\.gitconfig",
						"C:\\local\\config",
					},
				},
				{
					desc: "PREFIX should override system conf if set",
					env: env.NewFromKVList([]string{
						"HOME=C:\\home",
					}),
					cfg: &Config{
						Prefix:           "C\\prefix",
						LocalConfig:      "C:\\local\\config",
						SkipSystemConfig: false,
					},
					expectedOut: []string{
						"C\\prefix\\etc\\gitconfig",
						"C:\\home\\.gitconfig",
						"C:\\local\\config",
					},
				},
			}
			for i, tc := range testCases {
				tc := tc
				i := i
				t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
					t.Parallel()

					paths := getPaths(tc.env, tc.cfg)
					require.Equal(t, tc.expectedOut, paths)
				})
			}
		})
	default:
		t.Run("unix based OS", func(t *testing.T) {
			t.Parallel()
			testCases := []struct {
				desc        string
				env         *env.Env
				cfg         *Config
				expectedOut []string
			}{
				{
					desc: "No env and skip, should return the local file",
					env:  env.NewFromKVList([]string{}),
					cfg: &Config{
						LocalConfig:      "/local/path/config",
						SkipSystemConfig: true,
					},
					expectedOut: []string{"/local/path/config"},
				},
				{
					desc: "No env and no skip, should return the local and system",
					env:  env.NewFromKVList([]string{}),
					cfg: &Config{
						LocalConfig:      "/local/path/config",
						SkipSystemConfig: false,
					},
					expectedOut: []string{
						"/etc/gitconfig",
						"/usr/local/etc/gitconfig",
						"/opt/homebrew/etc/gitconfig",
						"/local/path/config",
					},
				},
				{
					desc: "if XDG_CONFIG_HOME is set, it should be used instead of HOME/.config",
					env: env.NewFromKVList([]string{
						"XDG_CONFIG_HOME=/xdg",
						"HOME=/home",
					}),
					cfg: &Config{
						LocalConfig:      "/local/path/config",
						SkipSystemConfig: false,
					},
					expectedOut: []string{
						"/etc/gitconfig",
						"/usr/local/etc/gitconfig",
						"/opt/homebrew/etc/gitconfig",
						"/xdg/git/.gitconfig",
						"/home/.gitconfig",
						"/local/path/config",
					},
				},
				{
					desc: "if XDG_CONFIG_HOME is NOT set, HOME/.config should be used instead",
					env: env.NewFromKVList([]string{
						"HOME=/home",
					}),
					cfg: &Config{
						LocalConfig:      "/local/path/config",
						SkipSystemConfig: false,
					},
					expectedOut: []string{
						"/etc/gitconfig",
						"/usr/local/etc/gitconfig",
						"/opt/homebrew/etc/gitconfig",
						"/home/.config/.git/.gitconfig",
						"/home/.gitconfig",
						"/local/path/config",
					},
				},
				{
					desc: "PREFIX should override system conf if set",
					env: env.NewFromKVList([]string{
						"HOME=/home",
					}),
					cfg: &Config{
						Prefix:           "/prefix",
						LocalConfig:      "/local/path/config",
						SkipSystemConfig: false,
					},
					expectedOut: []string{
						"/prefix/etc/gitconfig",
						"/home/.config/.git/.gitconfig",
						"/home/.gitconfig",
						"/local/path/config",
					},
				},
			}
			for i, tc := range testCases {
				tc := tc
				i := i
				t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
					t.Parallel()

					paths := getPaths(tc.env, tc.cfg)
					require.Equal(t, tc.expectedOut, paths)
				})
			}
		})
	}
}
