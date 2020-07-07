package git

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/require"
)

// TestInit creates a directory, runs Init(), and
// checks that every has been created
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
		require.Equal(t, d, r.projectPath)
		require.Equal(t, filepath.Join(d, DotGitPath), r.path)
	})
}

// // TestLoad runs Load(), and expects no error
// func TestLoad(t *testing.T) {
// 	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
// 	defer cleanup()

// 	r, err := LoadRepository(repoPath)
// 	require.NoError(t, err, "failed loading a repo")
// 	require.NotNil(t, r, "repository should not be nil")

// 	// assert returned repository
// 	require.Equal(t, repoPath, r.projectPath)
// 	require.Equal(t, filepath.Join(repoPath, DotGitPath), r.path)
// }

// func TestRepositoryGetObject(t *testing.T) {
// 	t.Parallel()

// 	t.Run("loose object", func(t *testing.T) {
// 		t.Parallel()

// 		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
// 		defer cleanup()

// 		r, err := LoadRepository(repoPath)
// 		require.NoError(t, err, "failed loading a repo")

// 		oid, err := plumbing.NewOidFromStr("b07e28976ac8972715598f390964d53cf4dbc1bd")
// 		require.NoError(t, err)

// 		obj, err := r.GetObject(oid)
// 		require.NoError(t, err)
// 		require.NotNil(t, obj)

// 		assert.Equal(t, oid, obj.ID)
// 		assert.Equal(t, object.TypeBlob, obj.Type())
// 		assert.Equal(t, "package packfile", string(obj.Bytes()[:16]))
// 	})

// 	t.Run("Object from packfile", func(t *testing.T) {
// 		t.Parallel()

// 		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
// 		defer cleanup()

// 		r, err := LoadRepository(repoPath)
// 		require.NoError(t, err, "failed loading a repo")

// 		oid, err := plumbing.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
// 		require.NoError(t, err)

// 		obj, err := r.GetObject(oid)
// 		require.NoError(t, err)
// 		require.NotNil(t, obj)

// 		assert.Equal(t, oid, obj.ID)
// 		assert.Equal(t, object.TypeCommit, obj.Type())
// 	})
// }

// func TestRepositoryNewBlob(t *testing.T) {
// 	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
// 	defer cleanup()

// 	r, err := LoadRepository(repoPath)
// 	require.NoError(t, err, "failed loading a repo")

// 	data := "abcdefghijklmnopqrstuvwxyz"
// 	blob, err := r.NewBlob([]byte(data))
// 	require.NoError(t, err)
// 	assert.NotEqual(t, plumbing.NullOid, blob.ID())
// 	assert.Equal(t, []byte(data), blob.Bytes())
// }
