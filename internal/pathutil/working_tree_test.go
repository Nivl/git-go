package pathutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/pathutil"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkingTreeFromPath(t *testing.T) {
	t.Parallel()

	t.Run(".git dir without head shouldn't be returned", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		err := os.MkdirAll(filepath.Join(path, ".git"), 0o755)
		require.NoError(t, err)

		finalPath := filepath.Join(path, "a", "b", "c")
		err = os.MkdirAll(finalPath, 0o755)
		require.NoError(t, err)

		_, err = pathutil.WorkingTreeFromPath(finalPath, ".git")
		require.Error(t, err)
		require.Equal(t, pathutil.ErrNoRepo, err)
	})

	t.Run(".git file should be returned", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		err := os.WriteFile(filepath.Join(path, ".git"), []byte(""), 0o644)
		require.NoError(t, err)

		finalPath := filepath.Join(path, "a", "b", "c")
		err = os.MkdirAll(finalPath, 0o755)
		require.NoError(t, err)

		p, err := pathutil.WorkingTreeFromPath(finalPath, ".git")
		require.NoError(t, err)
		assert.Equal(t, path, p)
	})

	t.Run(".git dir with head should be returned", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		err := os.MkdirAll(filepath.Join(path, ".git"), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(path, ".git", "HEAD"), []byte(""), 0o644)
		require.NoError(t, err)

		finalPath := filepath.Join(path, "a", "b", "c")
		err = os.MkdirAll(finalPath, 0o755)
		require.NoError(t, err)

		p, err := pathutil.WorkingTreeFromPath(finalPath, ".git")
		require.NoError(t, err)
		assert.Equal(t, path, p)
	})

	t.Run("no repo should return an error", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		finalPath := filepath.Join(path, "a", "b", "c")
		err := os.MkdirAll(finalPath, 0o755)
		require.NoError(t, err)

		_, err = pathutil.WorkingTreeFromPath(finalPath, ".git")
		require.Error(t, err)
		assert.ErrorIs(t, err, pathutil.ErrNoRepo)
	})
}

func TestWorkingTree(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, err := pathutil.WorkingTree(".git")
		require.NoError(t, err)
	})
}
