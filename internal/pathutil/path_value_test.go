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

func TestNewDirPathFlagWithDefault(t *testing.T) {
	t.Parallel()

	t.Run("single valid path should pass", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		p := pathutil.NewDirPathFlagWithDefault("/tmp")
		err := p.Set(path)
		assert.NoError(t, err)
		assert.Equal(t, path, p.String())
		assert.Equal(t, "path", p.Type())
	})

	t.Run("no path should use default", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		p := pathutil.NewDirPathFlagWithDefault(path)
		assert.Equal(t, path, p.String())
	})

	t.Run("invalid path should fail", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		p := pathutil.NewDirPathFlagWithDefault("/tmp")
		err := p.Set(filepath.Join(path, "doesn't exists"))
		assert.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("path should concat", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		finalPath := filepath.Join(path, "a", "b", "c")
		err := os.MkdirAll(finalPath, 0o755)
		require.NoError(t, err)

		p := pathutil.NewDirPathFlagWithDefault("/tmp")
		err = p.Set(path)
		assert.NoError(t, err)
		err = p.Set("a")
		assert.NoError(t, err)
		err = p.Set("b")
		assert.NoError(t, err)
		err = p.Set("c")
		assert.NoError(t, err)

		assert.Equal(t, finalPath, p.String())
	})

	t.Run("empty values should be ignored", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		finalPath := filepath.Join(path, "a", "b", "c")
		err := os.MkdirAll(finalPath, 0o755)
		require.NoError(t, err)

		p := pathutil.NewDirPathFlagWithDefault("/tmp")
		err = p.Set(path)
		assert.NoError(t, err)
		err = p.Set("a")
		assert.NoError(t, err)
		err = p.Set("")
		assert.NoError(t, err)
		err = p.Set("b")
		assert.NoError(t, err)
		err = p.Set("")
		assert.NoError(t, err)
		err = p.Set("c")
		assert.NoError(t, err)
		err = p.Set("")
		assert.NoError(t, err)

		assert.Equal(t, finalPath, p.String())
	})

	t.Run("absolute path should overwrite", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		fullPath := filepath.Join(path, "a", "b", "c")
		path2 := filepath.Join(path, "a", "b")
		err := os.MkdirAll(fullPath, 0o755)
		require.NoError(t, err)

		p := pathutil.NewDirPathFlagWithDefault("/tmp")
		err = p.Set(fullPath)
		assert.NoError(t, err)
		err = p.Set(path2)
		assert.NoError(t, err)

		assert.Equal(t, path2, p.String())
	})

	t.Run("should fail if path is a file", func(t *testing.T) {
		t.Parallel()

		f, cleanup := testutil.TempFile(t)
		t.Cleanup(cleanup)

		p := pathutil.NewDirPathFlagWithDefault("/tmp")
		err := p.Set(f.Name())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})
}

func TestNewFilePathFlagWithDefault(t *testing.T) {
	t.Parallel()

	t.Run("should fail if path is a directory", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		p := pathutil.NewFilePathFlagWithDefault("/tmp")
		err := p.Set(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "is a directory")
	})

	t.Run("should pass if path is a file", func(t *testing.T) {
		t.Parallel()

		f, cleanup := testutil.TempFile(t)
		t.Cleanup(cleanup)

		p := pathutil.NewFilePathFlagWithDefault("/tmp")
		err := p.Set(f.Name())
		require.NoError(t, err)
	})
}

func TestNewPathFlagWithDefault(t *testing.T) {
	t.Parallel()

	t.Run("should pass if path is a directory", func(t *testing.T) {
		t.Parallel()

		path, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		p := pathutil.NewPathFlagWithDefault("/tmp")
		err := p.Set(path)
		require.NoError(t, err)
	})

	t.Run("should pass if path is a file", func(t *testing.T) {
		t.Parallel()

		f, cleanup := testutil.TempFile(t)
		t.Cleanup(cleanup)

		p := pathutil.NewPathFlagWithDefault("/tmp")
		err := p.Set(f.Name())
		require.NoError(t, err)
	})
}
