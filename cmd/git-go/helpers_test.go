package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRepository(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)

	tmpPath, cleanup := testhelper.TempDir(t)
	t.Cleanup(cleanup)

	testCases := []struct {
		desc        string
		C           string
		expectError bool
	}{
		{
			desc: "A given path should be used",
			C:    repoPath,
		},
		{
			desc:        "Invalid path should return an error",
			C:           filepath.Join(tmpPath, "nope"),
			expectError: true,
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			cfg := &config{
				C: testhelper.NewStringValue(tc.C),
			}
			repo, err := loadRepository(cfg)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			t.Cleanup(func() {
				assert.NoError(t, repo.Close())
			})

			require.NoError(t, err)
			require.NotNil(t, repo)
		})
	}
}
