package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/Nivl/git-go/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInit creates a directory, runs Init(), and
// checks that every has been created
func TestInit(t *testing.T) {
	// Setup
	d, err := ioutil.TempDir("", t.Name()+"-")
	require.NoError(t, err)
	defer os.RemoveAll(d)

	// Run logic
	r, err := InitRepository(d)
	require.NoError(t, err, "failed creating a repo")

	// assert returned repository
	require.Equal(t, d, r.projectPath)
	require.Equal(t, filepath.Join(d, DotGitPath), r.path)

	// Assert data on disk
	dir := "dir"
	file := "file"
	checks := []struct {
		path string
		typ  string
	}{
		{path: DotGitPath, typ: dir},
		{path: filepath.Join(DotGitPath, BranchesPath), typ: dir},
		{path: filepath.Join(DotGitPath, ObjectsPath), typ: dir},
		{path: filepath.Join(DotGitPath, ObjectsInfoPath), typ: dir},
		{path: filepath.Join(DotGitPath, ObjectsPackPath), typ: dir},
		{path: filepath.Join(DotGitPath, RefsPath), typ: dir},
		{path: filepath.Join(DotGitPath, RefsTagsPath), typ: dir},
		{path: filepath.Join(DotGitPath, RefsHeadsPath), typ: dir},

		{path: filepath.Join(DotGitPath, ConfigPath), typ: file},
		{path: filepath.Join(DotGitPath, DescriptionPath), typ: file},
		{path: filepath.Join(DotGitPath, HEADPath), typ: file},
	}

	for _, check := range checks {
		fp := filepath.Join(d, check.path)
		s, err := os.Stat(fp)
		require.NoError(t, err, "%s should exist at %s", check.path, fp)

		if check.typ == dir {
			require.True(t, s.IsDir(), "%s should be a directory at %s", check.path, fp)
		}
	}
}

// TestLoad runs Load(), and expects no error
func TestLoad(t *testing.T) {
	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	defer cleanup()

	r, err := LoadRepository(repoPath)
	require.NoError(t, err, "failed loading a repo")
	require.NotNil(t, r, "repository should not be nil")

	// assert returned repository
	require.Equal(t, repoPath, r.projectPath)
	require.Equal(t, filepath.Join(repoPath, DotGitPath), r.path)
}

func TestGetObject(t *testing.T) {
	t.Parallel()

	t.Run("dangling object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()

		r, err := LoadRepository(repoPath)
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

		r, err := LoadRepository(repoPath)
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
