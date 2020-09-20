package packfile_test

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/ginternals/packfile"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestNewFromFile(t *testing.T) {
	t.Parallel()

	t.Run("valid packfile should pass", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		packFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.pack"
		packFilePath := filepath.Join(repoPath, gitpath.DotGitPath, gitpath.ObjectsPackPath, packFileName)
		pack, err := packfile.NewFromFile(packFilePath)
		require.NoError(t, err)
		assert.NotNil(t, pack)
		t.Cleanup(func() {
			require.NoError(t, pack.Close())
		})
	})

	t.Run("indexfile should fail", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		packFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.idx"
		packFilePath := filepath.Join(repoPath, gitpath.DotGitPath, gitpath.ObjectsPackPath, packFileName)
		pack, err := packfile.NewFromFile(packFilePath)
		require.Error(t, err)
		assert.True(t, xerrors.Is(err, packfile.ErrInvalidMagic))
		assert.Nil(t, pack)
	})
}

func TestGetObject(t *testing.T) {
	t.Parallel()

	t.Run("valid object should return an object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		packFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.pack"
		packFilePath := filepath.Join(repoPath, gitpath.DotGitPath, gitpath.ObjectsPackPath, packFileName)
		pack, err := packfile.NewFromFile(packFilePath)
		require.NoError(t, err)
		assert.NotNil(t, pack)
		t.Cleanup(func() {
			require.NoError(t, pack.Close())
		})

		// TODO(melvin): Test multiple parents
		t.Run("commit", func(t *testing.T) {
			t.Parallel()

			commitOid, err := ginternals.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
			require.NoError(t, err)
			o, err := pack.GetObject(commitOid)
			require.NoError(t, err)
			require.Equal(t, object.TypeCommit, o.Type())
			commit, err := o.AsCommit()
			require.NoError(t, err)
			require.Equal(t, commitOid, commit.ID())
			require.NotZero(t, commit.Author())
			require.NotZero(t, commit.Committer())

			require.Len(t, commit.ParentIDs(), 1)
			parentOid, err := ginternals.NewOidFromStr("f96f63e52cb8862b2c2d1a8b868229259c57854e")
			require.NoError(t, err)
			assert.Equal(t, parentOid, commit.ParentIDs()[0])

			assert.Equal(t, "build: switch to go module\n", commit.Message())
			assert.Equal(t, "Melvin Laplanche", commit.Author().Name)
			assert.Equal(t, "Melvin Laplanche", commit.Committer().Name)

			treeOid, err := ginternals.NewOidFromStr("c799e9129faae8d358e4b6de7813d6f970607893")
			require.NoError(t, err)
			assert.Equal(t, treeOid, commit.TreeID())
		})

		t.Run("blob", func(t *testing.T) {
			t.Parallel()

			blobOid, err := ginternals.NewOidFromStr("3f2f87160d5b4217125264310c22bcdad5b0d8bb")
			require.NoError(t, err)
			o, err := pack.GetObject(blobOid)
			require.NoError(t, err)
			require.Equal(t, object.TypeBlob, o.Type())

			blob := o.AsBlob()
			require.Equal(t, blobOid, blob.ID())
			assert.Equal(t, 207, blob.Size())
			assert.Equal(t, "# Binaries for programs and plugins", string(blob.Bytes()[:35]))
		})

		t.Run("tree", func(t *testing.T) {
			t.Parallel()

			treeOid, err := ginternals.NewOidFromStr("c799e9129faae8d358e4b6de7813d6f970607893")
			require.NoError(t, err)
			o, err := pack.GetObject(treeOid)
			require.NoError(t, err)
			require.Equal(t, object.TypeTree, o.Type())

			tree, err := o.AsTree()
			require.NoError(t, err)
			require.Equal(t, treeOid, tree.ID())
			require.Len(t, tree.Entries(), 18)

			// check a random entry
			entryOid, err := ginternals.NewOidFromStr("215559fe5053786726a19571fe0fd3d76c7fcfcd")
			require.NoError(t, err)
			entry := object.TreeEntry{
				Mode: 0o100644,
				ID:   entryOid,
				Path: "const.go",
			}
			require.Equal(t, entry, tree.Entries()[6])
		})

		t.Run("tag", func(t *testing.T) {
			t.Parallel()

			// TODO(melvin): we now support tags
			t.Skip("tags not yet supported")
		})
	})
}

func TestObjectCount(t *testing.T) {
	t.Parallel()

	t.Run("count the amount of objects in the test repo", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		// Load the packfile
		packFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.pack"
		packFilePath := filepath.Join(repoPath, gitpath.DotGitPath, gitpath.ObjectsPackPath, packFileName)
		pack, err := packfile.NewFromFile(packFilePath)
		require.NoError(t, err)
		assert.NotNil(t, pack)
		t.Cleanup(func() {
			require.NoError(t, pack.Close())
		})

		// TODO(melvin): this will break each time we change update
		// the test repo.
		// This needs to be rewritten once we have a way to create packfile
		assert.Equal(t, uint32(364), pack.ObjectCount())
	})
}
