package packfile_test

import (
	"bufio"
	"errors"
	"os"
	"testing"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/packfile"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/Nivl/git-go/internal/testutil/confutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIndex(t *testing.T) {
	t.Parallel()

	t.Run("valid indexfile should pass", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		indexFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.idx"
		cfg := confutil.NewCommonConfig(t, repoPath)
		indexFilePath := ginternals.PackfilePath(cfg, indexFileName)

		f, err := os.Open(indexFilePath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, f.Close())
		})

		index, err := packfile.NewIndex(bufio.NewReader(f))
		require.NoError(t, err)
		assert.NotNil(t, index)
	})

	t.Run("a packfile should fail", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		indexFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.pack"
		cfg := confutil.NewCommonConfig(t, repoPath)
		indexFilePath := ginternals.PackfilePath(cfg, indexFileName)

		f, err := os.Open(indexFilePath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, f.Close())
		})

		index, err := packfile.NewIndex(bufio.NewReader(f))
		require.Error(t, err)
		assert.Nil(t, index)
		assert.True(t, errors.Is(err, packfile.ErrInvalidMagic))
	})
}

func TestGetObjectOffset(t *testing.T) {
	t.Parallel()

	t.Run(string(testutil.RepoSmall), func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		indexFileName := "pack-0163931160835b1de2f120e1aa7e52206debeb14.idx"
		cfg := confutil.NewCommonConfig(t, repoPath)
		indexFilePath := ginternals.PackfilePath(cfg, indexFileName)

		f, err := os.Open(indexFilePath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, f.Close())
		})

		index, err := packfile.NewIndex(bufio.NewReader(f))
		require.NoError(t, err)
		assert.NotNil(t, index)

		t.Run("should work with valid oid", func(t *testing.T) {
			t.Parallel()

			oid, err := ginternals.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
			require.NoError(t, err)
			offset, err := index.GetObjectOffset(oid)
			require.NoError(t, err)
			assert.Equal(t, uint64(23081), offset)
		})

		t.Run("should fail with invalid oid", func(t *testing.T) {
			t.Parallel()

			oid, err := ginternals.NewOidFromStr("1acdadc2a420225783794fbffd51e2e137a69646")
			require.NoError(t, err)
			_, err = index.GetObjectOffset(oid)
			require.Error(t, err)
			require.True(t, errors.Is(err, ginternals.ErrObjectNotFound), "invalid error returned: %s", err.Error())
		})
	})
}
