package fsbackend

import (
	"os"
	"path/filepath"
	"testing"

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

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))
		obj, err := b.Object(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID())
		assert.Equal(t, object.TypeBlob, obj.Type())
		assert.Equal(t, "package packfile", string(obj.Bytes()[:16]))
	})

	t.Run("existing object in packfile should be returned", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		oid, err := plumbing.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))
		obj, err := b.Object(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID())
		assert.Equal(t, object.TypeCommit, obj.Type())
	})

	t.Run("un-existing object should fail", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		oid, err := plumbing.NewOidFromStr("2dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))
		obj, err := b.Object(oid)
		require.Error(t, err)
		require.Nil(t, obj)
		require.True(t, xerrors.Is(err, plumbing.ErrObjectNotFound), "unexpected error received")
	})
}

func TestHasObject(t *testing.T) {
	t.Run("existing object should exist", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))

		oid, err := plumbing.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		exists, err := b.HasObject(oid)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existing object should not exist", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))

		fakeOid, err := plumbing.NewOidFromStr("2dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		exists, err := b.HasObject(fakeOid)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("cache should be updated", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))

		oid, err := plumbing.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		_, found := b.cache.Get(oid)
		require.False(t, found, "the sha should have not been in the cache")

		exists, err := b.HasObject(oid)
		require.NoError(t, err)
		assert.True(t, exists, "the sha should exist")

		_, found = b.cache.Get(oid)
		require.True(t, found, "the sha should have been added to the cache")
	})
}

func TestWriteObject(t *testing.T) {
	t.Parallel()

	t.Run("add a new blob", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		dotGitPath := filepath.Join(repoPath, gitpath.DotGitPath)
		b := New(dotGitPath)
		o := object.New(object.TypeBlob, []byte("data"))
		oid, err := b.WriteObject(o)
		require.NoError(t, err)
		assert.NotEqual(t, plumbing.NullOid, oid, "invalid oid returned")

		// assert it's in disk
		storedO, err := b.Object(oid)
		require.NoError(t, err)
		assert.Equal(t, o.Type(), storedO.Type(), "invalid type")
		assert.Equal(t, o.Size(), storedO.Size(), "invalid size")
		assert.Equal(t, o.Bytes(), storedO.Bytes(), "invalid size")
		assert.NotEqual(t, plumbing.NullOid, storedO.ID(), "invalid ID")

		// make sure the blob was persisted
		p := filepath.Join(dotGitPath, gitpath.ObjectsPath, storedO.ID().String()[0:2], storedO.ID().String()[2:])
		_, err = os.Stat(p)
		require.NoError(t, err)
	})
}
