package fsbackend_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/backend/fsbackend"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestInit(t *testing.T) {
	t.Run("regular repo should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		err := fsbackend.New(filepath.Join(dir, gitpath.DotGitPath)).Init()
		require.NoError(t, err)
	})

	t.Run("bare repo should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		err := fsbackend.New(dir).Init()
		require.NoError(t, err)
	})

	t.Run("repo with existing data should work", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		// create a directory
		err := os.MkdirAll(filepath.Join(dir, gitpath.ObjectsPath), 0750)
		require.NoError(t, err)

		// create a file
		err = ioutil.WriteFile(filepath.Join(dir, gitpath.DescriptionPath), []byte{}, 0644)
		require.NoError(t, err)

		err = fsbackend.New(dir).Init()
		require.NoError(t, err)
	})

	t.Run("should fail if directory exists without write perm", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		// create a directory
		err := os.MkdirAll(filepath.Join(dir, gitpath.ObjectsPath), 0550)
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
		defer cleanup()

		// create a file
		err := ioutil.WriteFile(filepath.Join(dir, gitpath.DescriptionPath), []byte{}, 0444)
		require.NoError(t, err)

		err = fsbackend.New(dir).Init()
		require.Error(t, err)
		var perror *os.PathError
		require.True(t, xerrors.As(err, &perror), "error should be os.PathError")
		assert.Equal(t, "permission denied", perror.Err.Error())
	})
}
