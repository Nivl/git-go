package main

import (
	"fmt"
	"testing"

	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestLoadRepository(t *testing.T) {
	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)

	testCases := []struct {
		desc        string
		C           string
		expectError bool
	}{
		{
			desc: "no path should use the current directory",
			C:    "",
		},
		{
			desc: "A given path should be used",
			C:    repoPath,
		},
		{
			desc:        "Invalid path should return an error",
			C:           "/invalid/path",
			expectError: true,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			cfg := &config{
				C: tc.C,
			}
			repo, err := loadRepository(cfg)
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, repo)
		})
	}
}
