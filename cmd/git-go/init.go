package main

import (
	"fmt"

	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/spf13/cobra"
)

func newInitCmd(cfg *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init a new git repository",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return initCmd(cfg)
	}

	return cmd
}

func initCmd(cfg *globalFlags) error {
	p, err := config.LoadConfig(cfg.env, config.LoadConfigOptions{
		WorkingDirectory: cfg.C.String(),
		GitDirPath:       cfg.GitDir,
		WorkTreePath:     cfg.WorkTree,
		IsBare:           cfg.Bare,
		SkipGitDirLookUp: true,
	})
	if err != nil {
		return fmt.Errorf("could not create param: %w", err)
	}

	r, err := git.InitRepositoryWithParams(p, git.InitOptions{
		IsBare: cfg.Bare,
	})
	if err != nil {
		return err
	}
	return r.Close()
}
