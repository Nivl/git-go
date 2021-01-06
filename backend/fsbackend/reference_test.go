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
	t.Parallel()

	t.Run("Should fail if reference doesn't exists", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := New(filepath.Join(repoPath, gitpath.DotGitPath))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		ref, err := b.Reference("refs/heads/doesnt_exists")
		require.Error(t, err)
		assert.True(t, xerrors.Is(ginternals.ErrRefNotFound, ginternals.ErrRefNotFound), "unexpected error returned")
		assert.Nil(t, ref)
	})

	t.Run("Should success to follow a symbolic ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := New(filepath.Join(repoPath, gitpath.DotGitPath))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		ref, err := b.Reference(ginternals.Head)
		require.NoError(t, err)
		require.NotNil(t, ref)

		expectedTarget, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		assert.Equal(t, ginternals.Head, ref.Name())
		assert.Equal(t, "refs/heads/ml/packfile/tests", ref.SymbolicTarget())
		assert.Equal(t, expectedTarget, ref.Target())
	})

	t.Run("Should success to follow an oid ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := New(filepath.Join(repoPath, gitpath.DotGitPath))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		ref, err := b.Reference(gitpath.LocalBranch(ginternals.Master))
		require.NoError(t, err)
		require.NotNil(t, ref)

		expectedTarget, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		assert.Equal(t, gitpath.LocalBranch(ginternals.Master), ref.Name())
		assert.Empty(t, ref.SymbolicTarget())
		assert.Equal(t, expectedTarget, ref.Target())
	})
}

func TestParsePackedRefs(t *testing.T) {
	t.Parallel()

	createRepo := func(t *testing.T) (dir string, cleanup func()) {
		dir, cleanup = testhelper.TempDir(t)
		b, err := New(dir)
		require.NoError(t, err)
		defer require.NoError(t, b.Close())
		require.NoError(t, b.Init())
		return dir, cleanup
	}

	t.Run("Should return empty list if no files", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := createRepo(t)
		t.Cleanup(cleanup)

		b, err := New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		// By default it should only have HEAD
		assert.Len(t, b.refs, 1)
	})

	t.Run("Should fail if file contains invalid data", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := createRepo(t)
		t.Cleanup(cleanup)

		fPath := filepath.Join(dir, gitpath.PackedRefsPath)
		err := ioutil.WriteFile(fPath, []byte("not valid data"), 0o644)
		require.NoError(t, err)

		_, err = New(dir)
		require.Error(t, err)
		assert.True(t, xerrors.Is(err, ginternals.ErrPackedRefInvalid), "unexpected error received")
	})

	t.Run("Should pass with comments and annotations", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := createRepo(t)
		t.Cleanup(cleanup)

		fPath := filepath.Join(dir, gitpath.PackedRefsPath)
		err := ioutil.WriteFile(fPath, []byte("^de111c003b5661db802f17ac69419dcb9f4f3137\n# this is a comment"), 0o644)
		require.NoError(t, err)

		_, err = New(dir)
		require.NoError(t, err)
	})

	t.Run("Should correctly extract data", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := New(filepath.Join(repoPath, gitpath.DotGitPath))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, err)
		require.Len(t, b.refs, 14)
		expected := map[string][]byte{
			"HEAD": []byte("ref: refs/heads/ml/packfile/tests\n"),
			"FETCH_HEAD": []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089		branch 'master' of github.com:Nivl/git-go\n5f35f2dc6cec7356da02ca26192ce2bc3f271e79	not-for-merge	branch 'ml/feat/clone' of github.com:Nivl/git-go\n"),
			"ORIG_HEAD":                             []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089\n"),
			"refs/heads/master":                     []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089"),
			"refs/heads/ml/cleanup-062020":          []byte("b328320060eb503cf337c7cff281712ef236963a"),
			"refs/heads/ml/packfile/tests":          []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089"),
			"refs/heads/ml/tests":                   []byte("f0f70144f38695250606b86a50cff2b440a417f3"),
			"refs/remotes/origin/master":            []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089"),
			"refs/remotes/origin/ml/cleanup-062020": []byte("b328320060eb503cf337c7cff281712ef236963a"),
			"refs/remotes/origin/ml/feat/clone":     []byte("5f35f2dc6cec7356da02ca26192ce2bc3f271e79"),
			"refs/remotes/origin/HEAD":              []byte("ref: refs/remotes/origin/master\n"),
			"refs/stash":                            []byte("3fe6cf63fceced491a79fe634eb1e2c888225707"),
			"refs/tags/annotated":                   []byte("80316e01dbfdf5c2a8a20de66c747ecd4c4bd442\n"),
			"refs/tags/lightweight":                 []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089\n"),
		}
		assert.Equal(t, expected, b.refs)
	})
}

func TestWriteReference(t *testing.T) {
	t.Parallel()

	t.Run("should pass writing a new symbolic reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		b, err := New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init())

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
		t.Cleanup(cleanup)

		b, err := New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init())

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
		t.Cleanup(cleanup)

		b, err := New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init())

		ref := ginternals.NewSymbolicReference("H EAD", "refs/heads/master")
		err = b.WriteReference(ref)
		require.Error(t, err)
		require.True(t, xerrors.Is(err, ginternals.ErrRefNameInvalid), "unexpected error")
	})

	t.Run("should pass overwriting a symbolic reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := New(filepath.Join(repoPath, gitpath.DotGitPath))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

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
		t.Cleanup(cleanup)

		b, err := New(filepath.Join(repoPath, gitpath.DotGitPath))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

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
	t.Parallel()

	t.Run("should pass writing a new symbolic reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		b, err := New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init())

		ref := ginternals.NewSymbolicReference("refs/heads/my_feature", "refs/heads/master")
		err = b.WriteReferenceSafe(ref)
		require.NoError(t, err)

		// Let's make sure the data changed on disk
		data, err := ioutil.ReadFile(filepath.Join(b.root, "refs", "heads", "my_feature"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/master\n", string(data))
	})

	t.Run("should pass writing a new oid reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		b, err := New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init())

		target, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		ref := ginternals.NewReference("refs/heads/my_feature", target)
		err = b.WriteReferenceSafe(ref)
		require.NoError(t, err)

		// Let's make sure the data changed on disk
		data, err := ioutil.ReadFile(filepath.Join(b.root, "refs", "heads", "my_feature"))
		require.NoError(t, err)
		assert.Equal(t, target.String()+"\n", string(data))
	})

	t.Run("should fail with invalid name", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		b, err := New(dir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init())

		ref := ginternals.NewSymbolicReference("H EAD", "refs/heads/master")
		err = b.WriteReferenceSafe(ref)
		require.Error(t, err)
		require.True(t, xerrors.Is(err, ginternals.ErrRefNameInvalid), "unexpected error")
	})

	t.Run("should fail overwritting a ref on disk", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		b, err := New(filepath.Join(repoPath, gitpath.DotGitPath))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

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
		t.Cleanup(cleanup)

		b, err := New(filepath.Join(repoPath, gitpath.DotGitPath))
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		// assert current data on disk (there are none)
		_, err = ioutil.ReadFile(filepath.Join(b.root, "refs", "heads", "master"))
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
