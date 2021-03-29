package main

import (
	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/pathutil"
)

func loadRepository(cfg *config) (*git.Repository, error) {
	repoPath := cfg.C.String()
	root, err := pathutil.RepoRootFromPath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = root

	// run the command
	return git.OpenRepositoryWithOptions(repoPath, git.OpenOptions{
		GitOptions: newEnvOptionsFromCfg(cfg),
	})
}

func newEnvOptionsFromCfg(cfg *config) *env.GitOptions {
	return &env.GitOptions{
		GitDirPath:       cfg.GitDir,
		GitObjectDirPath: cfg.GitObjectDir,
		SkipSystemConfig: cfg.SkipSystemConfig,
		GitConfig:        cfg.GitConfig,
	}
}
