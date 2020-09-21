package fsbackend_test

import (
	"io/ioutil"
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

		err := fsbackend.New(filepath.Join(dir, gitpath.DotGitPath)).Init()
		require.NoError(t, err)
	})

	t.Run("bare repo should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		err := fsbackend.New(dir).Init()
		require.NoError(t, err)
	})

	t.Run("repo with existing data should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		// create a directory
		err := os.MkdirAll(filepath.Join(dir, gitpath.ObjectsPath), 0o750)
		require.NoError(t, err)

		// create a file
		err = ioutil.WriteFile(filepath.Join(dir, gitpath.DescriptionPath), []byte{}, 0o644)
		require.NoError(t, err)

		err = fsbackend.New(dir).Init()
		require.NoError(t, err)
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

		err = fsbackend.New(dir).Init()
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
		err := ioutil.WriteFile(filepath.Join(dir, gitpath.DescriptionPath), []byte{}, 0o444)
		require.NoError(t, err)

		err = fsbackend.New(dir).Init()
		require.Error(t, err)
		var perror *os.PathError
		require.True(t, xerrors.As(err, &perror), "error should be os.PathError")
		assert.Contains(t, perror.Err.Error(), "denied")
	})
}
