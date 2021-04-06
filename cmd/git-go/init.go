package main

import (
	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

func newInitCmd(cfg *flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init a new git repository",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return initCmd(cfg)
	}

	return cmd
}

func initCmd(cfg *flags) error {
	p, err := config.NewGitParams(cfg.env, config.NewGitParamsOptions{
		WorkingDirectory: cfg.C.String(),
		GitDirPath:       cfg.GitDir,
		WorkTreePath:     cfg.WorkTree,
		IsBare:           cfg.Bare,
		SkipGitDirLookUp: true,
	})
	if err != nil {
		return xerrors.Errorf("could not create param: %w", err)
	}

	r, err := git.InitRepositoryWithParams(p, git.InitOptions{
		IsBare: cfg.Bare,
	})
	if err != nil {
		return err
	}
	return r.Close()
}
