package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/env"
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

		err := initCmd(&config{C: &testhelper.StringValue{Value: dirPath}})
		require.NoError(t, err)

		info, err := os.Stat(filepath.Join(dirPath, ".git"))
		require.NoError(t, err)
		assert.True(t, info.IsDir(), "expected .git to be a dir")
	})

	t.Run("should fail on invalid path", func(t *testing.T) {
		t.Parallel()

		err := initCmd(&config{C: &testhelper.StringValue{Value: filepath.FromSlash("/this/path/is/fake")}})
		require.Error(t, err)
	})
}
