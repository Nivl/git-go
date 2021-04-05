package backend_test

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	t.Parallel()

	dir, cleanup := testhelper.TempDir(t)
	t.Cleanup(cleanup)

	dotGitPath := filepath.Join(dir, gitpath.DotGitPath)

	opts, err := config.NewGitOptionsSkipEnv(config.NewGitParamsOptions{
		WorkTreePath: dir,
		GitDirPath:   dotGitPath,
	})
	require.NoError(t, err)
	b, err := backend.NewFS(opts)
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

		opts, err := config.NewGitOptionsSkipEnv(config.NewGitParamsOptions{
			WorkTreePath: dir,
			GitDirPath:   dotGitPath,
		})
		require.NoError(t, err)
		b, err := backend.NewFS(opts)
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

		gitDirPath := filepath.Join(dir, gitpath.DotGitPath)
		objectDirPath := filepath.Join(dir, "objectDirPath")

		e := env.NewFromKVList([]string{
			"GIT_DIR=" + gitDirPath,
			"GIT_OBJECT_DIRECTORY=" + objectDirPath,
		})
		p, err := config.NewGitParams(e, config.NewGitParamsOptions{
			IsBare: true,
		})
		require.NoError(t, err)

		b, err := backend.NewFS(p)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.Equal(t, objectDirPath, b.ObjectsPath())
	})
}
