package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Run("repo with working tree", func(t *testing.T) {
		t.Parallel()

		// Setup
		d, cleanup := testhelper.TempDir(t)
		defer cleanup()

		// Run logic
		r, err := InitRepository(d)
		require.NoError(t, err, "failed creating a repo")

		// assert returned repository
		assert.Equal(t, d, r.repoRoot)
		assert.Equal(t, filepath.Join(d, gitpath.DotGitPath), r.dotGitPath)
		assert.NotNil(t, r.wt)
		assert.False(t, r.IsBare(), "repos should not be bare")
	})

	t.Run("bare repo", func(t *testing.T) {
		t.Parallel()

		// Setup
		d, cleanup := testhelper.TempDir(t)
		defer cleanup()

		// Run logic
		r, err := InitRepositoryWithOptions(d, InitOptions{
			IsBare: true,
		})
		require.NoError(t, err, "failed creating a repo")

		// assert returned repository
		require.Equal(t, d, r.repoRoot)
		require.Equal(t, d, r.dotGitPath)
		assert.Nil(t, r.wt)
		assert.True(t, r.IsBare(), "repos should be bare")
	})
}

func TestOpen(t *testing.T) {
	t.Run("repo with working tree", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")

		// assert returned repository
		assert.Equal(t, repoPath, r.repoRoot)
		assert.Equal(t, filepath.Join(repoPath, gitpath.DotGitPath), r.dotGitPath)
		assert.NotNil(t, r.wt)
		assert.False(t, r.IsBare(), "repos should not be bare")
	})

	t.Run("bare repo", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()
		repoPath = filepath.Join(repoPath, gitpath.DotGitPath)

		r, err := OpenRepositoryWithOptions(repoPath, OpenOptions{
			IsBare: true,
		})
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")

		// assert returned repository
		// assert returned repository
		require.Equal(t, repoPath, r.repoRoot)
		require.Equal(t, repoPath, r.dotGitPath)
		assert.Nil(t, r.wt)
		assert.True(t, r.IsBare(), "repos should be bare")
	})
}

func TestRepositoryGetObject(t *testing.T) {
	t.Parallel()

	t.Run("loose object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")

		oid, err := plumbing.NewOidFromStr("b07e28976ac8972715598f390964d53cf4dbc1bd")
		require.NoError(t, err)

		obj, err := r.GetObject(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID)
		assert.Equal(t, object.TypeBlob, obj.Type())
		assert.Equal(t, "package packfile", string(obj.Bytes()[:16]))
	})

	t.Run("Object from packfile", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")

		oid, err := plumbing.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		obj, err := r.GetObject(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID)
		assert.Equal(t, object.TypeCommit, obj.Type())
	})
}

func TestRepositoryNewBlob(t *testing.T) {
	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	defer cleanup()

	r, err := OpenRepository(repoPath)
	require.NoError(t, err, "failed loading a repo")

	data := "abcdefghijklmnopqrstuvwxyz"
	blob, err := r.NewBlob([]byte(data))
	require.NoError(t, err)
	assert.NotEqual(t, plumbing.NullOid, blob.ID())
	assert.Equal(t, []byte(data), blob.Bytes())

	// make sure the blob was persisted
	p := filepath.Join(r.dotGitPath, gitpath.ObjectsPath, blob.ID().String()[0:2], blob.ID().String()[2:])
	_, err = os.Stat(p)
	require.NoError(t, err)
}
