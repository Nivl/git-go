// Package confutil contains helpers and function to generate basic
// configuration
package confutil

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/ginternals/config"
	"github.com/stretchr/testify/require"
)

func NewCommonConfig(t *testing.T, workingTreePath string) *config.Config {
	cfg, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
		WorkTreePath: workingTreePath,
		GitDirPath:   filepath.Join(workingTreePath, config.DefaultDotGitDirName),
	})
	require.NoError(t, err)
	return cfg
}

func NewCommonConfigBare(t *testing.T, workingTreePath string) *config.Config {
	cfg, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
		IsBare:     true,
		GitDirPath: workingTreePath,
	})
	require.NoError(t, err)
	return cfg
}
