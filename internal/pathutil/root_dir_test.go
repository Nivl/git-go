package pathutil_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/pathutil"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoRootFromPath(t *testing.T) {
	t.Parallel()

	t.Run("subdir should be found", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		err := ioutil.WriteFile(filepath.Join(path, "HEAD"), []byte("ref: refs/heads/main"), 0o644)
		require.NoError(t, err)

		finalPath := filepath.Join(path, "a", "b", "c")
		err = os.MkdirAll(finalPath, 0o755)
		require.NoError(t, err)

		p, err := pathutil.RepoRootFromPath(finalPath)
		require.NoError(t, err)
		assert.Equal(t, path, p)
	})

	t.Run("bare repo should be found", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		err := os.MkdirAll(filepath.Join(path, ".git"), 0o755)
		require.NoError(t, err)

		finalPath := filepath.Join(path, "a", "b", "c")
		err = os.MkdirAll(finalPath, 0o755)
		require.NoError(t, err)

		p, err := pathutil.RepoRootFromPath(finalPath)
		require.NoError(t, err)
		assert.Equal(t, path, p)
	})

	t.Run("no repo should return an error", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		finalPath := filepath.Join(path, "a", "b", "c")
		err := os.MkdirAll(finalPath, 0o755)
		require.NoError(t, err)

		_, err = pathutil.RepoRootFromPath(finalPath)
		require.Error(t, err)
		assert.ErrorIs(t, err, pathutil.ErrNoRepo)
	})
}

func TestRepoRoot(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, err := pathutil.RepoRoot()
		require.NoError(t, err)
	})
}
