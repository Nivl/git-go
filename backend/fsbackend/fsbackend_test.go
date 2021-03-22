package fsbackend_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Nivl/git-go/backend/fsbackend"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestInit(t *testing.T) {
	t.Parallel()

	t.Run("regular repo should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		b, err := fsbackend.New(filepath.Join(dir, gitpath.DotGitPath))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, b.Init())
	})

	t.Run("repo with separated object dir", func(t *testing.T) {
		t.Parallel()

		repo, cleanupRepo := testhelper.TempDir(t)
		t.Cleanup(cleanupRepo)

		gitDir := filepath.Join(repo, gitpath.DotGitPath)
		objectDir := filepath.Join(repo, "git-objects")
		b, err := fsbackend.NewWithObjectsPath(gitDir, objectDir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, b.Init())

		_, err = os.Stat(gitDir)
		assert.NoError(t, err)

		// Check the directories that should exists
		_, err = os.Stat(gitDir)
		assert.NoError(t, err)
		_, err = os.Stat(objectDir)
		assert.NoError(t, err)
		_, err = os.Stat(filepath.Join(objectDir, gitpath.ObjectsInfoPath))
		assert.NoError(t, err)

		// Check the directories that should NOT exists
		_, err = os.Stat(filepath.Join(gitDir, gitpath.ObjectsPath))
		assert.Error(t, err)
	})

	t.Run("bare repo should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		b, err := fsbackend.New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, b.Init())
	})

	t.Run("repo with existing data should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		// create a directory
		err := os.MkdirAll(filepath.Join(dir, gitpath.ObjectsPath), 0o750)
		require.NoError(t, err)

		// create a file
		err = os.WriteFile(filepath.Join(dir, gitpath.DescriptionPath), []byte{}, 0o644)
		require.NoError(t, err)

		b, err := fsbackend.New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, b.Init())
	})

	t.Run("should fail if directory exists without write perm", func(t *testing.T) {
		t.Parallel()

		// TODO(melvin): Go to the bottom of this, somehow
		if runtime.GOOS == "windows" {
			t.Skip("Windows doesn't seem to be blocking writes.")
		}

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		// create a directory
		err := os.MkdirAll(filepath.Join(dir, gitpath.ObjectsPath), 0o550)
		require.NoError(t, err)

		b, err := fsbackend.New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		err = b.Init()
		require.Error(t, err)
		var perror *os.PathError
		require.True(t, xerrors.As(err, &perror), "error should be os.PathError")
		assert.Equal(t, "permission denied", perror.Err.Error())
	})

	t.Run("should fail if file exists without write perm", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		// create a file
		err := os.WriteFile(filepath.Join(dir, gitpath.DescriptionPath), []byte{}, 0o444)
		require.NoError(t, err)

		b, err := fsbackend.New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		err = b.Init()
		require.Error(t, err)
		var perror *os.PathError
		require.True(t, xerrors.As(err, &perror), "error should be os.PathError")
		assert.Contains(t, perror.Err.Error(), "denied")
	})
}
