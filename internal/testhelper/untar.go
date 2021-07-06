package testhelper

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/pathutil"
	"github.com/Nivl/git-go/internal/testhelper/exe"
	"github.com/stretchr/testify/require"
)

// RepoName represent the name of a test repository
type RepoName string

const (
	// RepoSmall is a snapshot of this repo up to commit bbb720a
	// from Fri Jun 19 18:16:17 2020 -0700
	RepoSmall RepoName = "small_repo"
)

// UnTar will untar a git repository in a new temporary folder.
func UnTar(t *testing.T, repoName RepoName) (repoPath string, cleanup func()) {
	repoPath, cleanup = TempDir(t)

	_, err := exe.Run("tar",
		"-xzf", fmt.Sprintf("%s/%s.tar.gz", TestdataPath(t), repoName),
		"-C", repoPath,
	)
	require.NoError(t, err)
	return repoPath, cleanup
}

// TestdataPath returns the absolute path to the testdata directory
func TestdataPath(t *testing.T) string {
	root, err := pathutil.WorkingTree(".git")
	require.NoError(t, err)
	return filepath.Join(root, "internal", "testdata")
}
