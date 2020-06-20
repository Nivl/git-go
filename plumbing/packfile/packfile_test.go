package packfile_test

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/plumbing/packfile"
	"github.com/Nivl/git-go/plumbing/packfile/mockpackfile"
	"github.com/Nivl/git-go/testhelper"
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
