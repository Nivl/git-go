package backend

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/ginternals/packfile"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestObject(t *testing.T) {
	t.Parallel()

	t.Run("existing loose object should be returned", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		oid, err := ginternals.NewOidFromStr("b07e28976ac8972715598f390964d53cf4dbc1bd")
		require.NoError(t, err)

		b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
			ProjectPath: repoPath,
		}))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

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
		t.Cleanup(cleanup)

		oid, err := ginternals.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
			ProjectPath: repoPath,
		}))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		obj, err := b.Object(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID())
		assert.Equal(t, object.TypeCommit, obj.Type())
	})

	t.Run("un-existing object should fail", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		oid, err := ginternals.NewOidFromStr("2dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
			ProjectPath: repoPath,
		}))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		obj, err := b.Object(oid)
		require.Error(t, err)
		require.Nil(t, obj)
		require.True(t, xerrors.Is(err, ginternals.ErrObjectNotFound), "unexpected error received")
	})
}

func TestHasObject(t *testing.T) {
	t.Parallel()

	t.Run("existing object should exist", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
			ProjectPath: repoPath,
		}))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		oid, err := ginternals.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		exists, err := b.HasObject(oid)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existing object should not exist", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
			ProjectPath: repoPath,
		}))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		fakeOid, err := ginternals.NewOidFromStr("2dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		exists, err := b.HasObject(fakeOid)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("cache should be updated", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
			ProjectPath: repoPath,
		}))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		oid, err := ginternals.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		_, found := b.cache.Get(oid)
		require.False(t, found, "the sha should have not been in the cache")

		exists, err := b.HasObject(oid)
		require.NoError(t, err)
		assert.True(t, exists, "the sha should exist")

		_, found = b.cache.Get(oid)
		require.True(t, found, "the sha should have been added to the cache")

		// should get the data from the cache
		exists, err = b.HasObject(oid)
		require.NoError(t, err)
		assert.True(t, exists, "the sha should exist")
	})

	t.Run("invalid cache should be replaced", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
			ProjectPath: repoPath,
		}))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		oid, err := ginternals.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		b.cache.Add(oid, "not a valid value")

		exists, err := b.HasObject(oid)
		require.NoError(t, err)
		assert.True(t, exists, "the sha should exist")

		o, found := b.cache.Get(oid)
		require.True(t, found, "the sha should have been added to the cache")
		require.IsType(t, &object.Object{}, o)
	})
}

func TestWriteObject(t *testing.T) {
	t.Parallel()

	t.Run("add a new blob", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
			ProjectPath: repoPath,
		}))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		o := object.New(object.TypeBlob, []byte("data"))
		oid, err := b.WriteObject(o)
		require.NoError(t, err)
		assert.NotEqual(t, ginternals.NullOid, oid, "invalid oid returned")

		// assert it's in disk
		storedO, err := b.Object(oid)
		require.NoError(t, err)
		assert.Equal(t, o.Type(), storedO.Type(), "invalid type")
		assert.Equal(t, o.Size(), storedO.Size(), "invalid size")
		assert.Equal(t, o.Bytes(), storedO.Bytes(), "invalid size")
		assert.NotEqual(t, ginternals.NullOid, storedO.ID(), "invalid ID")

		// make sure the blob was persisted
		p := filepath.Join(b.ObjectsPath(), storedO.ID().String()[0:2], storedO.ID().String()[2:])
		info, err := os.Stat(p)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o444), info.Mode(), "objects should be read only")
	})

	t.Run("Writing the same object twice should not trigger a rewrite", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
			ProjectPath: repoPath,
		}))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		o := object.New(object.TypeBlob, []byte("data"))
		oid, err := b.WriteObject(o)
		require.NoError(t, err)
		assert.NotEqual(t, ginternals.NullOid, oid, "invalid oid returned")

		// assert it's on the disk
		storedO, err := b.Object(oid)
		require.NoError(t, err)
		assert.Equal(t, o.Type(), storedO.Type(), "invalid type")
		assert.Equal(t, o.Size(), storedO.Size(), "invalid size")
		assert.Equal(t, o.Bytes(), storedO.Bytes(), "invalid size")
		assert.NotEqual(t, ginternals.NullOid, storedO.ID(), "invalid ID")

		// make sure the blob was persisted
		p := filepath.Join(b.ObjectsPath(), storedO.ID().String()[0:2], storedO.ID().String()[2:])
		originalInfo, err := os.Stat(p)
		require.NoError(t, err)

		// let's rewrite the same file
		time.Sleep(10 * time.Millisecond)
		_, err = b.WriteObject(o)
		require.NoError(t, err)
		info, err := os.Stat(p)
		require.NoError(t, err)

		assert.Equal(t, originalInfo.ModTime(), info.ModTime())
	})
}

func TestWalkPackedObjectIDs(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)
	b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
		ProjectPath: repoPath,
	}))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, b.Close())
	})

	t.Run("Should return all the objects", func(t *testing.T) {
		t.Parallel()

		totalObject := 0
		err := b.WalkPackedObjectIDs(func(oid ginternals.Oid) error {
			totalObject++
			return nil
		})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, totalObject, 100)
	})

	t.Run("Should stop the walk", func(t *testing.T) {
		t.Parallel()

		totalObject := 0
		err := b.WalkPackedObjectIDs(func(oid ginternals.Oid) error {
			if totalObject == 4 {
				return packfile.OidWalkStop
			}
			totalObject++
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 4, totalObject)
	})

	t.Run("Should propage an error", func(t *testing.T) {
		t.Parallel()

		someErr := errors.New("some error")
		totalObject := 0
		err := b.WalkPackedObjectIDs(func(oid ginternals.Oid) error {
			if totalObject == 4 {
				return someErr
			}
			totalObject++
			return nil
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, someErr)
		assert.Equal(t, 4, totalObject)
	})
}

func TestLoosePackedObjectIDs(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)
	b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
		ProjectPath: repoPath,
	}))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, b.Close())
	})

	t.Run("Should return all the objects", func(t *testing.T) {
		t.Parallel()

		totalObject := 0
		err := b.WalkLooseObjectIDs(func(oid ginternals.Oid) error {
			totalObject++
			return nil
		})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, totalObject, 2)
	})

	t.Run("Should stop the walk", func(t *testing.T) {
		t.Parallel()

		totalObject := 0
		err := b.WalkLooseObjectIDs(func(oid ginternals.Oid) error {
			totalObject++
			return packfile.OidWalkStop
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, totalObject)
	})

	t.Run("Should propage an error", func(t *testing.T) {
		t.Parallel()

		someErr := errors.New("some error")
		err := b.WalkLooseObjectIDs(func(oid ginternals.Oid) error {
			return someErr
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, someErr)
	})
}

func TestIsLooseObjectDir(t *testing.T) {
	t.Parallel()

	dir, cleanup := testhelper.TempDir(t)
	t.Cleanup(cleanup)

	b, err := NewFS(env.NewDefaultGitOptions(env.FinalizeOptions{
		ProjectPath: dir,
	}))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, b.Close())
	})

	t.Run("Any directory from 00 to ff should be valid", func(t *testing.T) {
		t.Parallel()

		for i := int64(0); i < 256; i++ {
			hex := fmt.Sprintf("%02x", 255)
			assert.True(t, b.isLooseObjectDir(hex), "%s (%d) should pass", hex, i)
		}
	})

	shouldFail := true
	testCases := []struct {
		desc     string
		name     string
		expected bool
	}{
		{
			desc:     "Should fail with a name too long",
			name:     "fff",
			expected: shouldFail,
		},
		{
			desc:     "Should fail with a name too short",
			name:     "f",
			expected: shouldFail,
		},
		{
			desc:     "Should fail with an invalid hex",
			name:     "gg",
			expected: shouldFail,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, !b.isLooseObjectDir(tc.name), tc.expected)
		})
	}
}
