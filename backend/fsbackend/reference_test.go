package fsbackend_test

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/backend/fsbackend"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/Nivl/git-go/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestReference(t *testing.T) {
	t.Run("Should fail if reference doesn't exists", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := fsbackend.New(filepath.Join(repoPath, gitpath.DotGitPath))
		ref, err := b.Reference("refs/heads/doesnt_exists")
		require.Error(t, err)
		assert.True(t, xerrors.Is(plumbing.ErrRefNotFound, plumbing.ErrRefNotFound), "unexpected error returned")
		assert.Nil(t, ref)
	})

	t.Run("Should success to follow a symbolic ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := fsbackend.New(filepath.Join(repoPath, gitpath.DotGitPath))
		ref, err := b.Reference(plumbing.HEAD)
		require.NoError(t, err)
		require.NotNil(t, ref)

		expectedTarget := plumbing.NullOid
		assert.Equal(t, plumbing.HEAD, ref.Name())
		assert.Equal(t, "refs/remotes/origin/HEAD", ref.SymbolicTarget())
		assert.Equal(t, expectedTarget, ref.Target())
	})

	t.Run("Should success to follow an oid ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := fsbackend.New(filepath.Join(repoPath, gitpath.DotGitPath))
		ref, err := b.Reference(plumbing.MasterLocalRef)
		require.NoError(t, err)
		require.NotNil(t, ref)

		expectedTarget, err := plumbing.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		assert.Equal(t, plumbing.MasterLocalRef, ref.Name())
		assert.Empty(t, ref.SymbolicTarget())
		assert.Equal(t, expectedTarget, ref.Target())
	})
}
