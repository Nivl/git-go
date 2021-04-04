package env

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDotGitPath(t *testing.T) {
	t.Parallel()

	// To be able to build an absolute path on Windows we need to know
	// the Volume name
	dir, err := os.Getwd()
	require.NoError(t, err)
	root := filepath.VolumeName(dir) + string(os.PathSeparator)

	testCases := []struct {
		desc      string
		repoPath  string
		gitDirCfg string
		isBare    bool
		expected  string
	}{
		{
			desc:      "Test basic repo",
			repoPath:  filepath.Join(root, "path", "to", "repo"),
			gitDirCfg: "",
			isBare:    false,
			expected:  filepath.Join(root, "path", "to", "repo", gitpath.DotGitPath),
		},
		{
			desc:      "Test bare repo",
			repoPath:  filepath.Join(root, "path", "to", "repo"),
			gitDirCfg: "",
			isBare:    true,
			expected:  filepath.Join(root, "path", "to", "repo"),
		},
		{
			desc:      "Test repo with absolute config path",
			repoPath:  filepath.Join(root, "path", "to", "working-tree"),
			gitDirCfg: filepath.Join(root, "path", "to", "repo"),
			isBare:    false,
			expected:  filepath.Join(root, "path", "to", "repo"),
		},
		{
			desc:      "Test repo with relative config path",
			repoPath:  filepath.Join(root, "path", "to", "working-tree"),
			gitDirCfg: filepath.Join("repo"),
			isBare:    false,
			expected:  filepath.Join(root, "path", "to", "working-tree", "repo"),
		},
		{
			desc:      "Test bare repo with relative config path",
			repoPath:  filepath.Join(root, "path", "to", "working-tree"),
			gitDirCfg: filepath.Join("repo"),
			isBare:    true,
			expected:  filepath.Join(root, "path", "to", "working-tree", "repo"),
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			opts := &GitOptions{
				GitDirPath: tc.gitDirCfg,
			}
			out := opts.buildDotGitPath(tc.repoPath, tc.isBare)
			assert.Equal(t, tc.expected, out)
		})
	}
}

func TestBuildDotGitObjectsPat(t *testing.T) {
	t.Parallel()

	// To be able to build an absolute path on Windows we need to know
	// the Volume name
	dir, err := os.Getwd()
	require.NoError(t, err)
	root := filepath.VolumeName(dir) + string(os.PathSeparator)

	testCases := []struct {
		desc           string
		repoPath       string
		dotGitPath     string
		objectsPathCfg string
		expected       string
	}{
		{
			desc:           "Test basic repo",
			repoPath:       filepath.Join(root, "path", "to", "repo"),
			dotGitPath:     filepath.Join(root, "path", "to", "repo", gitpath.DotGitPath),
			objectsPathCfg: "",
			expected:       filepath.Join(root, "path", "to", "repo", gitpath.DotGitPath, gitpath.ObjectsPath),
		},
		{
			desc:           "Test repo with absolute config path",
			repoPath:       filepath.Join(root, "path", "to", "repo"),
			dotGitPath:     filepath.Join(root, "path", "to", "repo", gitpath.DotGitPath),
			objectsPathCfg: filepath.Join(root, "path", "to", "objects"),
			expected:       filepath.Join(root, "path", "to", "objects"),
		},
		{
			desc:           "Test repo with relative config path",
			repoPath:       filepath.Join(root, "path", "to", "repo"),
			dotGitPath:     filepath.Join(root, "path", "to", "repo", gitpath.DotGitPath),
			objectsPathCfg: filepath.Join("objects"),
			expected:       filepath.Join(root, "path", "to", "repo", "objects"),
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()
			opts := &GitOptions{
				GitObjectDirPath: tc.objectsPathCfg,
			}
			out := opts.buildDotGitObjectsPath(tc.repoPath, tc.dotGitPath)
			assert.Equal(t, tc.expected, out)
		})
	}
}
