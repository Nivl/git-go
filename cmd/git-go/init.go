package main

import (
	"github.com/Nivl/git-go"
	"github.com/spf13/cobra"
)

func newInitCmd(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init a new git repository",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return initCmd(cfg)
	}

	return cmd
}

func initCmd(cfg *config) error {
	r, err := git.InitRepositoryWithOptions(cfg.C.String(), git.InitOptions{
		EnvOptions: newEnvOptionsFromCfg(cfg),
	})
	if err != nil {
		return err
	}
	return r.Close()
}
