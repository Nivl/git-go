package ginternals_test

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/stretchr/testify/require"
)

func TestLocalTagFullName(t *testing.T) {
	t.Parallel()

	out := ginternals.LocalTagFullName("my-tag/nested")
	expect := "refs/tags/my-tag/nested"
	require.Equal(t, expect, out)
}

func TestLocalTagShortName(t *testing.T) {
	t.Parallel()

	out := ginternals.LocalTagShortName("refs/tags/my-tag/nested")
	expect := "my-tag/nested"
	require.Equal(t, expect, out)
}

func TestLocalBranchFullName(t *testing.T) {
	t.Parallel()

	out := ginternals.LocalBranchFullName("my-branch/nested")
	expect := "refs/heads/my-branch/nested"
	require.Equal(t, expect, out)
}

func TestLocalBranchShortName(t *testing.T) {
	t.Parallel()

	out := ginternals.LocalBranchShortName("refs/heads/my-branch/nested")
	expect := "my-branch/nested"
	require.Equal(t, expect, out)
}

func TestRefFullName(t *testing.T) {
	t.Parallel()

	out := ginternals.RefFullName("HEAD")
	expect := "refs/HEAD"
	require.Equal(t, expect, out)
}

func TestRefsPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		CommonDirPath: "common",
	}

	out := ginternals.RefsPath(cfg)
	expect := filepath.Join("common", "refs")
	require.Equal(t, expect, out)
}

func TestRefPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		CommonDirPath: "common",
	}

	out := ginternals.RefPath(cfg, "heads/main")
	expect := filepath.Join("common", "refs", "heads", "main")
	require.Equal(t, expect, out)
}

func TestPackedRefsPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		CommonDirPath: "common",
	}

	out := ginternals.PackedRefsPath(cfg)
	expect := filepath.Join("common", "packed-refs")
	require.Equal(t, expect, out)
}

func TestTagsPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		CommonDirPath: "common",
	}

	out := ginternals.TagsPath(cfg)
	expect := filepath.Join("common", "refs", "tags")
	require.Equal(t, expect, out)
}

func TestDotGitPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		GitDirPath: ".git",
	}

	out := ginternals.DotGitPath(cfg)
	expect := ".git"
	require.Equal(t, expect, out)
}

func TestLocalBranchesPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		CommonDirPath: "common",
	}

	out := ginternals.LocalBranchesPath(cfg)
	expect := filepath.Join("common", "refs", "heads")
	require.Equal(t, expect, out)
}

func TestObjectsPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ObjectDirPath: "objects",
	}

	out := ginternals.ObjectsPath(cfg)
	expect := "objects"
	require.Equal(t, expect, out)
}

func TestObjectsInfoPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ObjectDirPath: "objects",
	}

	out := ginternals.ObjectsInfoPath(cfg)
	expect := filepath.Join("objects", "info")
	require.Equal(t, expect, out)
}

func TestObjectsPacksPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ObjectDirPath: "objects",
	}

	out := ginternals.ObjectsPacksPath(cfg)
	expect := filepath.Join("objects", "pack")
	require.Equal(t, expect, out)
}

func TestPackfilePath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ObjectDirPath: "objects",
	}

	out := ginternals.PackfilePath(cfg, "my_pack.pack")
	expect := filepath.Join("objects", "pack", "my_pack.pack")
	require.Equal(t, expect, out)
}

func TestConfigPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LocalConfig: "config",
	}

	out := ginternals.ConfigPath(cfg)
	expect := "config"
	require.Equal(t, expect, out)
}

func TestDescriptionFilePath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		GitDirPath: ".git",
	}

	out := ginternals.DescriptionFilePath(cfg)
	expect := filepath.Join(".git", "description")
	require.Equal(t, expect, out)
}

func TestLooseObjectPath(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		ObjectDirPath: "objects",
	}

	out := ginternals.LooseObjectPath(cfg, "fcfe68a0e44e04bd7fd564fc0b75f1ae457e18b3")
	expect := filepath.Join("objects", "fc", "fe68a0e44e04bd7fd564fc0b75f1ae457e18b3")
	require.Equal(t, expect, out)
}
