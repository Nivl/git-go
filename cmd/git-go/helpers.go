package main

import (
	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals/config"
	"golang.org/x/xerrors"
)

func loadRepository(cfg *flags) (*git.Repository, error) {
	p, err := config.NewGitParams(cfg.env, config.NewGitParamsOptions{
		WorkingDirectory: cfg.C.String(),
		GitDirPath:       cfg.GitDir,
		WorkTreePath:     cfg.WorkTree,
		IsBare:           cfg.Bare,
	})
	if err != nil {
		return nil, xerrors.Errorf("could not create param: %w", err)
	}

	// run the command
	return git.OpenRepositoryWithParams(p, git.OpenOptions{
		IsBare: cfg.Bare,
	})
}
