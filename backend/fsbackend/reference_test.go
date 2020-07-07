package fsbackend

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))
		ref, err := b.Reference("refs/heads/doesnt_exists")
		require.Error(t, err)
		assert.True(t, xerrors.Is(plumbing.ErrRefNotFound, plumbing.ErrRefNotFound), "unexpected error returned")
		assert.Nil(t, ref)
	})

	t.Run("Should success to follow a symbolic ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))
		ref, err := b.Reference(plumbing.HEAD)
		require.NoError(t, err)
		require.NotNil(t, ref)

		expectedTarget, err := plumbing.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		assert.Equal(t, plumbing.HEAD, ref.Name())
		assert.Equal(t, "refs/heads/ml/packfile/tests", ref.SymbolicTarget())
		assert.Equal(t, expectedTarget, ref.Target())
	})

	t.Run("Should success to follow an oid ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))
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

func TestParsePackedRefs(t *testing.T) {
	t.Run("Should return empty list if no files", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "fsbackend-init-")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		b := New(dir)
		err = b.Init()
		require.NoError(t, err)

		data, err := b.parsePackedRefs()
		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Empty(t, data)
	})

	t.Run("Should fail if file contains invalid data", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "fsbackend-init-")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		b := New(dir)
		err = b.Init()
		require.NoError(t, err)
		fPath := filepath.Join(b.root, gitpath.PackedRefsPath)
		err = ioutil.WriteFile(fPath, []byte("not valid data"), 0644)
		require.NoError(t, err)

		_, err = b.parsePackedRefs()
		require.Error(t, err)
		assert.True(t, xerrors.Is(err, plumbing.ErrPackedRefInvalid), "unexpected error received")
	})

	t.Run("Should pass with comments and annotations", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "fsbackend-init-")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		b := New(dir)
		err = b.Init()
		require.NoError(t, err)
		fPath := filepath.Join(b.root, gitpath.PackedRefsPath)
		err = ioutil.WriteFile(fPath, []byte("^de111c003b5661db802f17ac69419dcb9f4f3137\n# this is a comment"), 0644)
		require.NoError(t, err)

		_, err = b.parsePackedRefs()
		require.NoError(t, err)
	})

	t.Run("Should correctly extract data", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))

		data, err := b.parsePackedRefs()
		require.NoError(t, err)
		require.Len(t, data, 8)
		expected := map[string]string{
			"refs/heads/master":                     "bbb720a96e4c29b9950a4c577c98470a4d5dd089",
			"refs/heads/ml/cleanup-062020":          "b328320060eb503cf337c7cff281712ef236963a",
			"refs/heads/ml/packfile/tests":          "bbb720a96e4c29b9950a4c577c98470a4d5dd089",
			"refs/heads/ml/tests":                   "f0f70144f38695250606b86a50cff2b440a417f3",
			"refs/remotes/origin/master":            "bbb720a96e4c29b9950a4c577c98470a4d5dd089",
			"refs/remotes/origin/ml/cleanup-062020": "b328320060eb503cf337c7cff281712ef236963a",
			"refs/remotes/origin/ml/feat/clone":     "5f35f2dc6cec7356da02ca26192ce2bc3f271e79",
			"refs/stash":                            "3fe6cf63fceced491a79fe634eb1e2c888225707",
		}
		assert.Equal(t, expected, data)
	})
}
