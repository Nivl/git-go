package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/testhelper"
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
