package testhelper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	out, err := ioutil.TempDir("", strings.ReplaceAll(t.Name(), "/", "_")+"_")
	require.NoError(t, err)

	defer func() {
		if err != nil {
			os.RemoveAll(out) //nolint
		}
	}()

	_, err = exe.Run("tar",
		"-xzf", fmt.Sprintf("%s/%s.tar.gz", TestdataPath(t), repoName),
		"-C", out,
	)
	require.NoError(t, err)
	return out, func() {
		require.NoError(t, os.RemoveAll(out))
	}
}

// TestdataPath returns the absolute path to the testdata directory
func TestdataPath(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err)
	projectRootName := "git-go"
	projectRootIndex := strings.Index(wd, projectRootName)
	if projectRootIndex == -1 {
		require.FailNow(t, "could not find project root")
	}
	return filepath.Join(wd[:projectRootIndex], projectRootName, "internal", "testdata")
}
