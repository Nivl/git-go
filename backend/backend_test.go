package backend_test

import (
	"testing"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/Nivl/git-go/internal/testutil/confutil"
	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	t.Parallel()

	dir, cleanup := testutil.TempDir(t)
	t.Cleanup(cleanup)

	cfg := confutil.NewCommonConfig(t, dir)
	b, err := backend.NewFS(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, b.Close())
	})

	require.Equal(t, cfg.GitDirPath, b.Path())
}
