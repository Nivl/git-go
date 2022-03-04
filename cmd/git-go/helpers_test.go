package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRepository(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
	t.Cleanup(cleanup)

	tmpPath, cleanup := testutil.TempDir(t)
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

			cfg := &globalFlags{
				env: env.NewFromKVList([]string{}),
				C:   testutil.NewStringValue(tc.C),
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
