package main

import (
	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/Nivl/git-go/internal/pathutil"
	"golang.org/x/xerrors"
)

func loadRepository(cfg *flags) (*git.Repository, error) {
	repoPath := cfg.C.String()
	root, err := pathutil.RepoRootFromPath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = root

	opts, err := config.NewGitOptions(cfg.env, config.NewGitOptionsParams{
		ProjectPath: repoPath,
	})
	if err != nil {
		return nil, xerrors.Errorf("could not create options: %w", err)
	}

	// run the command
	return git.OpenRepositoryWithOptions(repoPath, git.OpenOptions{
		GitOptions: opts,
	})
}
