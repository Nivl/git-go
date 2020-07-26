package fsbackend

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
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
		assert.True(t, xerrors.Is(ginternals.ErrRefNotFound, ginternals.ErrRefNotFound), "unexpected error returned")
		assert.Nil(t, ref)
	})

	t.Run("Should success to follow a symbolic ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))
		ref, err := b.Reference(ginternals.HEAD)
		require.NoError(t, err)
		require.NotNil(t, ref)

		expectedTarget, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		assert.Equal(t, ginternals.HEAD, ref.Name())
		assert.Equal(t, "refs/heads/ml/packfile/tests", ref.SymbolicTarget())
		assert.Equal(t, expectedTarget, ref.Target())
	})

	t.Run("Should success to follow an oid ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))
		ref, err := b.Reference(ginternals.MasterLocalRef)
		require.NoError(t, err)
		require.NotNil(t, ref)

		expectedTarget, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		assert.Equal(t, ginternals.MasterLocalRef, ref.Name())
		assert.Empty(t, ref.SymbolicTarget())
		assert.Equal(t, expectedTarget, ref.Target())
	})
}

func TestParsePackedRefs(t *testing.T) {
	t.Run("Should return empty list if no files", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		b := New(dir)
		err := b.Init()
		require.NoError(t, err)

		data, err := b.parsePackedRefs()
		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Empty(t, data)
	})

	t.Run("Should fail if file contains invalid data", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		b := New(dir)
		err := b.Init()
		require.NoError(t, err)
		fPath := filepath.Join(b.root, gitpath.PackedRefsPath)
		err = ioutil.WriteFile(fPath, []byte("not valid data"), 0o644)
		require.NoError(t, err)

		_, err = b.parsePackedRefs()
		require.Error(t, err)
		assert.True(t, xerrors.Is(err, ginternals.ErrPackedRefInvalid), "unexpected error received")
	})

	t.Run("Should pass with comments and annotations", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		b := New(dir)
		err := b.Init()
		require.NoError(t, err)
		fPath := filepath.Join(b.root, gitpath.PackedRefsPath)
		err = ioutil.WriteFile(fPath, []byte("^de111c003b5661db802f17ac69419dcb9f4f3137\n# this is a comment"), 0o644)
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

func TestWriteReference(t *testing.T) {
	t.Run("should pass writing a new symbolic reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		b := New(dir)
		err := b.Init()
		require.NoError(t, err)

		ref := ginternals.NewSymbolicReference("HEAD", "refs/heads/master")
		err = b.WriteReference(ref)
		require.NoError(t, err)

		data, err := ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/master\n", string(data))
	})

	t.Run("should pass writing a new oid reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		b := New(dir)
		err := b.Init()
		require.NoError(t, err)

		target, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		ref := ginternals.NewReference("HEAD", target)
		err = b.WriteReference(ref)
		require.NoError(t, err)

		data, err := ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, target.String()+"\n", string(data))
	})

	t.Run("should fail with invalid name", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		b := New(dir)
		err := b.Init()
		require.NoError(t, err)

		ref := ginternals.NewSymbolicReference("H EAD", "refs/heads/master")
		err = b.WriteReference(ref)
		require.Error(t, err)
		require.True(t, xerrors.Is(err, ginternals.ErrRefNameInvalid), "unexpected error")

		_, err = ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.Error(t, err)
	})

	t.Run("should pass overwriting a symbolic reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))

		// assert current data on disk
		data, err := ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/ml/packfile/tests\n", string(data))

		ref := ginternals.NewSymbolicReference("HEAD", "refs/heads/master")
		err = b.WriteReference(ref)
		require.NoError(t, err)

		data, err = ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/master\n", string(data))
	})

	t.Run("should pass overwriting an oid reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))

		// assert current data on disk
		data, err := ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/ml/packfile/tests\n", string(data))

		target, err := ginternals.NewOidFromStr("abb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		ref := ginternals.NewReference("HEAD", target)
		err = b.WriteReference(ref)
		require.NoError(t, err)

		data, err = ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, target.String()+"\n", string(data))
	})
}

func TestWriteReferenceSafe(t *testing.T) {
	t.Run("should pass writing a new symbolic reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		b := New(dir)
		err := b.Init()
		require.NoError(t, err)

		ref := ginternals.NewSymbolicReference("HEAD", "refs/heads/master")
		err = b.WriteReferenceSafe(ref)
		require.NoError(t, err)

		// Let's make sure the data changed on disk
		data, err := ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/master\n", string(data))
	})

	t.Run("should pass writing a new oid reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		b := New(dir)
		err := b.Init()
		require.NoError(t, err)

		target, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		ref := ginternals.NewReference("HEAD", target)
		err = b.WriteReferenceSafe(ref)
		require.NoError(t, err)

		// Let's make sure the data changed on disk
		data, err := ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, target.String()+"\n", string(data))
	})

	t.Run("should fail with invalid name", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		defer cleanup()

		b := New(dir)
		err := b.Init()
		require.NoError(t, err)

		ref := ginternals.NewSymbolicReference("H EAD", "refs/heads/master")
		err = b.WriteReferenceSafe(ref)
		require.Error(t, err)
		require.True(t, xerrors.Is(err, ginternals.ErrRefNameInvalid), "unexpected error")

		// Let's make sure the data have not been persisted
		_, err = ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.Error(t, err)
	})

	t.Run("should fail overwritting a ref on disk", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))

		// assert current data on disk
		data, err := ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/ml/packfile/tests\n", string(data))

		ref := ginternals.NewSymbolicReference("HEAD", "refs/heads/master")
		err = b.WriteReferenceSafe(ref)
		require.Error(t, err)
		require.True(t, xerrors.Is(err, ginternals.ErrRefExists), "unexpected error")

		// let's make sure the data have not changed
		data, err = ioutil.ReadFile(filepath.Join(b.root, "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/ml/packfile/tests\n", string(data))
	})

	t.Run("should fail overwritting a packed ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		b := New(filepath.Join(repoPath, gitpath.DotGitPath))

		// assert current data on disk (there are none)
		_, err := ioutil.ReadFile(filepath.Join(b.root, "refs", "heads", "master"))
		require.Error(t, err)

		ref := ginternals.NewSymbolicReference("refs/heads/master", "refs/heads/branch")
		err = b.WriteReferenceSafe(ref)
		require.Error(t, err)
		require.True(t, xerrors.Is(err, ginternals.ErrRefExists), "unexpected error")

		// Let's make sure the data have not been persisted
		_, err = ioutil.ReadFile(filepath.Join(b.root, "refs", "heads", "master"))
		require.Error(t, err)
	})
}
