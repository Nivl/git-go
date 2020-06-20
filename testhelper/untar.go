package testhelper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nivl/git-go/testhelper/exe"
	"github.com/stretchr/testify/require"
)

type RepoName string

const (
	RepoSmall RepoName = "small_repo"
)

func UnTar(t *testing.T, repoName RepoName) (string, func()) {
	out, err := ioutil.TempDir("", strings.ReplaceAll(t.Name(), "/", "_")+"_")
	require.NoError(t, err)

	defer func() {
		if err != nil {
			os.RemoveAll(out) //nolint
		}
	}()

	wd, err := os.Getwd()
	require.NoError(t, err)
	projectRootName := "git-go"
	projectRootIndex := strings.LastIndex(wd, projectRootName)
	if projectRootIndex == -1 {
		require.FailNow(t, "could not find project root")
	}
	dataDir := filepath.Join(wd[:projectRootIndex], projectRootName, "testdata")

	_, err = exe.Run("tar",
		"-xzf", fmt.Sprintf("%s/%s.tar.gz", dataDir, repoName),
		"-C", out,
	)
	require.NoError(t, err)
	return out, func() {
		require.NoError(t, os.RemoveAll(out))
	}
}
