package backend_test

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	t.Parallel()

	dir, cleanup := testhelper.TempDir(t)
	t.Cleanup(cleanup)

	dotGitPath := filepath.Join(dir, gitpath.DotGitPath)

	b, err := backend.NewFS(nil, dotGitPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, b.Close())
	})

	require.Equal(t, dotGitPath, b.Path())
}

func TestObjectPath(t *testing.T) {
	t.Parallel()

	t.Run("automatically set on dotGit path", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		dotGitPath := filepath.Join(dir, gitpath.DotGitPath)

		b, err := backend.NewFS(nil, dotGitPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.Equal(t, filepath.Join(dotGitPath, gitpath.ObjectsPath), b.ObjectsPath())
	})

	t.Run("manually set", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		dotGitPath := filepath.Join(dir, gitpath.DotGitPath)
		objectsPath := filepath.Join(dir, "git-objects")

		b, err := backend.NewFSWithObjectsPath(nil, dotGitPath, objectsPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.Equal(t, objectsPath, b.ObjectsPath())
	})
}
