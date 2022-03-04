package backend_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/Nivl/git-go/internal/testutil/confutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Parallel()

	t.Run("regular repo should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := backend.NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, b.Init(ginternals.Master))

		data, err := os.ReadFile(filepath.Join(ginternals.DotGitPath(cfg), ginternals.Head))
		require.NoError(t, err)
		require.Equal(t, "ref: refs/heads/master\n", string(data))
	})

	t.Run("repo with separated object dir", func(t *testing.T) {
		t.Parallel()

		repo, cleanupRepo := testutil.TempDir(t)
		t.Cleanup(cleanupRepo)

		gitDirPath := filepath.Join(repo, ".git")
		objectDirPath := filepath.Join(repo, "git-objects")

		e := env.NewFromKVList([]string{
			"GIT_DIR=" + gitDirPath,
			"GIT_OBJECT_DIRECTORY=" + objectDirPath,
		})
		cfg, err := config.LoadConfig(e, config.LoadConfigOptions{
			IsBare: true,
		})
		require.NoError(t, err)

		b, err := backend.NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, b.Init(ginternals.Master))

		// Check the directories that should exists
		assert.DirExists(t, gitDirPath)
		assert.DirExists(t, objectDirPath)
		assert.DirExists(t, ginternals.ObjectsInfoPath(cfg))

		// Check the directories that should NOT exists
		assert.NoDirExists(t, filepath.Join(gitDirPath, "objects"))
	})

	t.Run("bare repo should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfigBare(t, dir)
		b, err := backend.NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, b.Init(ginternals.Master))
		assert.NoDirExists(t, filepath.Join(dir, config.DefaultDotGitDirName))
		assert.FileExists(t, filepath.Join(dir, ginternals.Head))
	})

	t.Run("repo with existing data should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		// create a directory
		err := os.MkdirAll(filepath.Join(dir, "objects"), 0o750)
		require.NoError(t, err)

		// create a file
		err = os.WriteFile(filepath.Join(dir, "description"), []byte{}, 0o644)
		require.NoError(t, err)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := backend.NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, b.Init(ginternals.Master))
	})

	t.Run("should fail if directory exists without write perm", func(t *testing.T) {
		t.Parallel()

		// TODO(melvin): Go to the bottom of this, somehow
		if runtime.GOOS == "windows" {
			t.Skip("Windows doesn't seem to be blocking writes.")
		}

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		// create a directory
		err := os.MkdirAll(filepath.Join(dir, "objects"), 0o550)
		require.NoError(t, err)

		cfg := confutil.NewCommonConfigBare(t, dir)
		b, err := backend.NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		err = b.Init(ginternals.Master)
		require.Error(t, err)
		var perror *os.PathError
		require.True(t, errors.As(err, &perror), "error should be os.PathError")
		assert.Equal(t, "permission denied", perror.Err.Error())
	})

	t.Run("should fail if file exists without write perm", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		// create a file
		err := os.WriteFile(filepath.Join(dir, "description"), []byte{}, 0o444)
		require.NoError(t, err)

		cfg := confutil.NewCommonConfigBare(t, dir)
		b, err := backend.NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		err = b.Init(ginternals.Master)
		require.Error(t, err)
		var perror *os.PathError
		require.True(t, errors.As(err, &perror), "error should be os.PathError")
		assert.Contains(t, perror.Err.Error(), "denied")
	})

	t.Run("should create a symlink", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
			WorkTreePath: filepath.Join(dir),
			GitDirPath:   filepath.Join(dir, "separate-dir"),
		})
		require.NoError(t, err)

		b, err := backend.NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, b.InitWithOptions(ginternals.Master, backend.InitOptions{
			CreateSymlink: true,
		}))

		gitfilePath := filepath.Join(dir, config.DefaultDotGitDirName)
		require.FileExists(t, gitfilePath)

		data, err := os.ReadFile(gitfilePath)
		require.NoError(t, err)

		expectedContent := "gitdir: " + filepath.Join(dir, "separate-dir")
		require.Equal(t, expectedContent, string(data))
	})
}
