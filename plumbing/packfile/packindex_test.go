package packfile_test

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/packfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestNewIndexFromFile(t *testing.T) {
	t.Parallel()

	t.Run("valid indexfile should pass", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		indexFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.idx"
		indexFilePath := filepath.Join(repoPath, git.DotGitPath, git.ObjectsPackPath, indexFileName)
		index, err := packfile.NewIndexFromFile(indexFilePath)
		require.NoError(t, err)
		assert.NotNil(t, index)
		defer func() {
			require.NoError(t, index.Close())
		}()
	})

	t.Run("a packfile should fail", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		indexFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.pack"
		indexFilePath := filepath.Join(repoPath, git.DotGitPath, git.ObjectsPackPath, indexFileName)
		index, err := packfile.NewIndexFromFile(indexFilePath)
		require.Error(t, err)
		assert.Nil(t, index)
		assert.True(t, xerrors.Is(err, packfile.ErrInvalidMagic))
	})
}

func TestGetObjectOffset(t *testing.T) {
	t.Parallel()

	t.Run(string(testhelper.RepoSmall), func(t *testing.T) {
		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		indexFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.idx"
		indexFilePath := filepath.Join(repoPath, git.DotGitPath, git.ObjectsPackPath, indexFileName)
		index, err := packfile.NewIndexFromFile(indexFilePath)
		require.NoError(t, err)
		assert.NotNil(t, index)
		defer func() {
			require.NoError(t, index.Close())
		}()

		t.Run("should work with valid oid", func(t *testing.T) {
			oid, err := plumbing.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
			require.NoError(t, err)
			offset, err := index.GetObjectOffset(oid)
			require.NoError(t, err)
			assert.Equal(t, uint64(23081), offset)
		})

		t.Run("should fail with invalid oid", func(t *testing.T) {
			oid, err := plumbing.NewOidFromStr("1acdadc2a420225783794fbffd51e2e137a69646")
			require.NoError(t, err)
			_, err = index.GetObjectOffset(oid)
			require.Error(t, err)
			require.True(t, xerrors.Is(err, plumbing.ErrObjectNotFound))
		})
	})
}
