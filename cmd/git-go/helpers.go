package main

import (
	"fmt"

	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals/config"
)

func loadRepository(cfg *flags) (*git.Repository, error) {
	p, err := config.LoadConfig(cfg.env, config.LoadConfigOptions{
		WorkingDirectory: cfg.C.String(),
		GitDirPath:       cfg.GitDir,
		WorkTreePath:     cfg.WorkTree,
		IsBare:           cfg.Bare,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create param: %w", err)
	}

	// run the command
	return git.OpenRepositoryWithParams(p, git.OpenOptions{
		IsBare: cfg.Bare,
	})
}
