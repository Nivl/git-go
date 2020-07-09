package fsbackend_test

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/backend/fsbackend"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestObject(t *testing.T) {
	t.Parallel()

	t.Run("existing loose object should be returned", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		oid, err := plumbing.NewOidFromStr("b07e28976ac8972715598f390964d53cf4dbc1bd")
		require.NoError(t, err)

		b := fsbackend.New(filepath.Join(repoPath, gitpath.DotGitPath))
		obj, err := b.Object(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID)
		assert.Equal(t, object.TypeBlob, obj.Type())
		assert.Equal(t, "package packfile", string(obj.Bytes()[:16]))
	})

	t.Run("existing object in packfile should be returned", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		oid, err := plumbing.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		b := fsbackend.New(filepath.Join(repoPath, gitpath.DotGitPath))
		obj, err := b.Object(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID)
		assert.Equal(t, object.TypeCommit, obj.Type())
	})

	t.Run("un-existing object should fail", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		oid, err := plumbing.NewOidFromStr("2dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		b := fsbackend.New(filepath.Join(repoPath, gitpath.DotGitPath))
		obj, err := b.Object(oid)
		require.Error(t, err)
		require.Nil(t, obj)
		require.True(t, xerrors.Is(err, plumbing.ErrObjectNotFound), "unexpected error received")
	})
}

func TestHasObject(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	defer cleanup()

	oid, err := plumbing.NewOidFromStr("2dcdadc2a420225783794fbffd51e2e137a69646")
	require.NoError(t, err)
	b := fsbackend.New(filepath.Join(repoPath, gitpath.DotGitPath))

	require.Panics(t, func() {
		b.HasObject(oid)
	})
}

func TestWriteObject(t *testing.T) {
	t.Parallel()

	t.Run("add a new blob", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := fsbackend.New(filepath.Join(repoPath, gitpath.DotGitPath))
		o := object.New(object.TypeBlob, []byte("data"))
		oid, err := b.WriteObject(o)
		require.NoError(t, err)

		// assert it's in disk
		storedO, err := b.Object(oid)
		require.NoError(t, err)
		assert.Equal(t, o.Type(), storedO.Type(), "invalid type")
		assert.Equal(t, o.Size(), storedO.Size(), "invalid size")
		assert.Equal(t, o.Bytes(), storedO.Bytes(), "invalid size")
	})
}
