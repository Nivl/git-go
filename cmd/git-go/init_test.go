package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitParams(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		args []string
	}{
		{
			desc: "should work with no options",
			args: []string{"init"},
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			dirPath, cleanup := testhelper.TempDir(t)
			t.Cleanup(cleanup)
			tc.args = append(tc.args, "-C", dirPath)

			cwd, err := os.Getwd()
			require.NoError(t, err)

			cmd := newRootCmd(cwd, env.NewFromOs())
			cmd.SetArgs(tc.args)

			require.NotPanics(t, func() {
				err = cmd.Execute()
			})
			require.NoError(t, err)
		})
	}
}

func TestInit(t *testing.T) {
	t.Parallel()

	t.Run("should work with default params", func(t *testing.T) {
		t.Parallel()

		dirPath, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		sdtout := bytes.NewBufferString("")

		err := initCmd(sdtout, &globalFlags{
			env: env.NewFromKVList([]string{}),
			C:   &testhelper.StringValue{Value: dirPath},
		}, initCmdFlags{})
		require.NoError(t, err)

		gitDir := filepath.Join(dirPath, config.DefaultDotGitDirName)
		info, err := os.Stat(gitDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir(), "expected .git to be a dir")

		expectedOut := fmt.Sprintf("Initialized empty Git repository in %s\n", gitDir)
		assert.Equal(t, expectedOut, sdtout.String())
	})

	t.Run("should create un-existing path", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		err := initCmd(io.Discard, &globalFlags{
			env: env.NewFromKVList([]string{}),
			C:   &testhelper.StringValue{Value: filepath.Join(dir, "this", "path", "is", "fake")},
		}, initCmdFlags{})
		require.NoError(t, err)
	})

	t.Run("should allow a branch name", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		err := initCmd(io.Discard,
			&globalFlags{
				env: env.NewFromKVList([]string{}),
				C:   &testhelper.StringValue{Value: dir},
			},
			initCmdFlags{
				initialBranch: "main",
			})
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(dir, config.DefaultDotGitDirName, ginternals.Head))
		require.NoError(t, err)
		require.Equal(t, "ref: refs/heads/main\n", string(data))
	})

	t.Run("Quiet should prevent writing data to stdout", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		sdtout := bytes.NewBufferString("")

		err := initCmd(sdtout,
			&globalFlags{
				env: env.NewFromKVList([]string{}),
				C:   &testhelper.StringValue{Value: dir},
			},
			initCmdFlags{
				quiet: true,
			})
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(dir, config.DefaultDotGitDirName, ginternals.Head))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/master\n", string(data))

		assert.Empty(t, sdtout.String(), "no output was expected")
	})
}
