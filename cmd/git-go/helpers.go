package main

import (
	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/internal/pathutil"
)

func loadRepository(cfg *config) (*git.Repository, error) {
	repoPath := cfg.C
	if repoPath == "" {
		root, err := pathutil.RepoRoot()
		if err != nil {
			return nil, err
		}
		repoPath = root
	}

	// run the command
	return git.LoadRepository(repoPath)
}