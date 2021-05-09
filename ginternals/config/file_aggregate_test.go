package config

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/Nivl/git-go/env"
	"github.com/stretchr/testify/require"
)

// func TestNewFileAggregate(t *testing.T) {
// 	t.Parallel()

// 	t.Run("common os", func(t *testing.T) {
// 		t.Parallel()
// 		testCases := []struct {
// 			desc          string
// 			env           *env.Env
// 			cfg           *Config
// 			expectedOut   *FileAggregate
// 			expectedError error
// 		}{
// 			{
// 				desc:        "should work with no env set",
// 				env:         env.NewFromKVList([]string{}),
// 				cfg:         &Config{},
// 				expectedOut: &FileAggregate{},
// 			},
// 		}
// 		for i, tc := range testCases {
// 			tc := tc
// 			i := i
// 			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
// 				t.Parallel()

// 				f, err := NewFileAggregate(tc.env, tc.cfg)
// 				switch tc.expectedError {
// 				default:
// 					require.Error(t, err)
// 					require.ErrorIs(t, err, tc.expectedError, "unexpected error")
// 					require.Nil(t, f)
// 				case nil:
// 					require.NoError(t, err)
// 					require.NotNil(t, f)
// 				}
// 			})
// 		}
// 	})
// }

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
						"C:\\local\\config",
						"C:\\home\\.gitconfig",
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

/**
t.Run("should fail creating a repo if the target dir is a file", func(t *testing.T) {
	t.Parallel()

	file, cleanup := testhelper.TempFile(t)
	t.Cleanup(cleanup)

	opts, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
		WorkingDirectory: file.Name(),
		SkipGitDirLookUp: true,
	})
	require.NoError(t, err)

	// Run logic
	_, err = InitRepositoryWithParams(opts, InitOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a directory")
})

t.Run("should fail creating a repo if one of the parent dir is a file", func(t *testing.T) {
	t.Parallel()

	// Setup
	file, cleanup := testhelper.TempFile(t)
	t.Cleanup(cleanup)

	opts, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
		WorkingDirectory: filepath.Join(file.Name(), "wt"),
		SkipGitDirLookUp: true,
	})
	require.NoError(t, err)

	// Run logic
	_, err = InitRepositoryWithParams(opts, InitOptions{})
	require.Error(t, err)

	// Windows seems to be ok to run a stat on a path that
	// contains a file as directory, but won't let you create
	// directories in that path.
	// UNIX will make the stat fail.
	switch runtime.GOOS {
	case "windows":
		require.Contains(t, err.Error(), "could not create")
	default:
		require.Contains(t, err.Error(), "could not check")
	}
})
**/
