package packfile_test

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/internal/mocks/mockpackfile"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/Nivl/git-go/plumbing/packfile"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestNewFromFile(t *testing.T) {
	t.Parallel()

	t.Run("valid packfile should pass", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		mockctrl := gomock.NewController(t)
		defer mockctrl.Finish()
		mockGetter := mockpackfile.NewMockObjectGetter(mockctrl)

		packFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.pack"
		packFilePath := filepath.Join(repoPath, git.DotGitPath, git.ObjectsPackPath, packFileName)
		pack, err := packfile.NewFromFile(mockGetter, packFilePath)
		require.NoError(t, err)
		assert.NotNil(t, pack)
		defer func() {
			require.NoError(t, pack.Close())
		}()
	})

	t.Run("indexfile should fail", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		mockctrl := gomock.NewController(t)
		defer mockctrl.Finish()
		mockGetter := mockpackfile.NewMockObjectGetter(mockctrl)

		packFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.idx"
		packFilePath := filepath.Join(repoPath, git.DotGitPath, git.ObjectsPackPath, packFileName)
		pack, err := packfile.NewFromFile(mockGetter, packFilePath)
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
		defer cleanup()

		mockctrl := gomock.NewController(t)
		defer mockctrl.Finish()
		mockGetter := mockpackfile.NewMockObjectGetter(mockctrl)

		packFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.pack"
		packFilePath := filepath.Join(repoPath, git.DotGitPath, git.ObjectsPackPath, packFileName)
		pack, err := packfile.NewFromFile(mockGetter, packFilePath)
		require.NoError(t, err)
		assert.NotNil(t, pack)
		defer func() {
			require.NoError(t, pack.Close())
		}()

		// TODO(melvin): Test multiple parents
		t.Run("commit", func(t *testing.T) {
			commitOid, err := plumbing.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
			require.NoError(t, err)
			o, err := pack.GetObject(commitOid)
			require.NoError(t, err)
			require.Equal(t, object.TypeCommit, o.Type())
			commit, err := o.AsCommit()
			require.NoError(t, err)
			require.Equal(t, commitOid, commit.ID)
			require.NotNil(t, commit.Author)
			require.NotNil(t, commit.Committer)

			require.Len(t, commit.ParentIDs, 1)
			parentOid, err := plumbing.NewOidFromStr("f96f63e52cb8862b2c2d1a8b868229259c57854e")
			require.NoError(t, err)
			assert.Equal(t, parentOid, commit.ParentIDs[0])

			assert.Equal(t, "build: switch to go module\n", commit.Message)
			assert.Equal(t, "Melvin Laplanche", commit.Author.Name)
			assert.Equal(t, "Melvin Laplanche", commit.Committer.Name)

			treeOid, err := plumbing.NewOidFromStr("c799e9129faae8d358e4b6de7813d6f970607893")
			require.NoError(t, err)
			assert.Equal(t, treeOid, commit.TreeID)
		})

		t.Run("blob", func(t *testing.T) {
			blobOid, err := plumbing.NewOidFromStr("3f2f87160d5b4217125264310c22bcdad5b0d8bb")
			require.NoError(t, err)
			o, err := pack.GetObject(blobOid)
			require.NoError(t, err)
			require.Equal(t, object.TypeBlob, o.Type())

			blob := o.AsBlob()
			require.Equal(t, blobOid, blob.ID)
			assert.Equal(t, 207, blob.Size())
			assert.Equal(t, "# Binaries for programs and plugins", string(blob.Bytes()[:35]))
		})

		t.Run("tree", func(t *testing.T) {
			treeOid, err := plumbing.NewOidFromStr("c799e9129faae8d358e4b6de7813d6f970607893")
			require.NoError(t, err)
			o, err := pack.GetObject(treeOid)
			require.NoError(t, err)
			require.Equal(t, object.TypeTree, o.Type())

			// TODO(melvin): Implement o.AsTree()
			// tree, err := o.AsTree()
			// require.NoError(t, err)
			// require.Equal(t, treeOid, tree.ID)

			// ❯ git cat-file -p c799e9129faae8d358e4b6de7813d6f970607893
			// 100644 blob 3f2f87160d5b4217125264310c22bcdad5b0d8bb	.gitignore
			// 100644 blob 0aab040a4e9cacd927497cd0649b8aa840dc3e97	README.md
			// 100644 blob db46468fdc3a0b749b162ec3c0b4b0a8f60fc024	blob.go
			// 040000 tree 6a025d914dc8325aef2b1b380ff8a1faa66839cf	cmd
			// 100644 blob 60058f39beb075aae92817161f5423e1d094f884	commit.go
			// 100644 blob be483a92271d1b632d32ef94d426c4cd995e0135	commit_test.go
			// 100644 blob 215559fe5053786726a19571fe0fd3d76c7fcfcd	const.go
			// 100644 blob 8ff6962cd658d55b11ebb79d2132a4560174206b	go.mod
			// 100644 blob 36cdb23c1e871b8fc08449d761f0132be007c92e	go.sum
			// 100644 blob 3351bc5f36b888b87d484d10fdbbccce21f7460d	object.go
			// 100644 blob ae561f3b6dbdfd77ec39d38e10736d4c51a0a692	object_test.go
			// 100644 blob 44b9cd4054f7c9ca2e020b9220496e10a5a658bf	oid.go
			// 100644 blob 680ddd7b6e091187b0fe6ec3abd3c09689507556	packfile.go
			// 100644 blob 964c34cec7d3b4dbce9ffd6aeed925e78f93a39c	packindex.go
			// 100644 blob 2fdef1dd4a9e5f7639d328bdef4cc28f25735d88	readutil.go
			// 100644 blob 191d2cfeb5e62ece55869517d0a06347b9eb13cd	repo.go
			// 100644 blob a4878edbf96c9ad4c57c4d576fa8d306267a9fc0	repo_test.go
			// 040000 tree 859a70442a1e4c23ea85a922530616d802d39523	vendor
		})

		t.Run("tag", func(t *testing.T) {
			t.Skip("tags not yet supported")
		})
	})
}
