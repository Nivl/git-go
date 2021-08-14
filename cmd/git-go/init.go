package main

import (
	"fmt"

	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/spf13/cobra"
)

// initCmdFlags represents the flags accepted by the init command
type initCmdFlags struct {
	initialBranch string
}

func newInitCmd(cfg *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init a new git repository",
	}

	flags := initCmdFlags{}
	cmd.Flags().StringVarP(&flags.initialBranch, "initial-branch", "b", "", "Use the specified name for the initial branch in the newly created repository. If not specified, fall back to the default name (currently master, but this is subject to change in the future; the name can be customized via the init.defaultBranch configuration variable).")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return initCmd(cfg, flags)
	}

	return cmd
}

func initCmd(cfg *globalFlags, flags initCmdFlags) error {
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
		IsBare:            cfg.Bare,
		InitialBranchName: flags.initialBranch,
	})
	if err != nil {
		return err
	}
	return r.Close()
}
