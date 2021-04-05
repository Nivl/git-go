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
	opts, err := config.NewGitOptions(cfg.env, config.NewGitOptionsParams{
		ProjectPath: cfg.C.String(),
	})
	if err != nil {
		return xerrors.Errorf("could not create options: %w", err)
	}

	r, err := git.InitRepositoryWithOptions(cfg.C.String(), git.InitOptions{
		GitOptions: opts,
	})
	if err != nil {
		return err
	}
	return r.Close()
}
