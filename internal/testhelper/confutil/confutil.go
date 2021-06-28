// Package confutil contains helpers and function to generate basic
// configuration
package confutil

import (
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/ginternals/config"
	"github.com/stretchr/testify/require"
)

// NewCommonConfig creates a new basic config object using the most common options
func NewCommonConfig(t *testing.T, workingTreePath string) *config.Config {
	cfg, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
		WorkTreePath: workingTreePath,
		GitDirPath:   filepath.Join(workingTreePath, config.DefaultDotGitDirName),
	})
	require.NoError(t, err)
	return cfg
}

// NewCommonConfigBare creates a new basic config object using the most common options
// for a bare repository
func NewCommonConfigBare(t *testing.T, workingTreePath string) *config.Config {
	cfg, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
		IsBare:     true,
		GitDirPath: workingTreePath,
	})
	require.NoError(t, err)
	return cfg
}
