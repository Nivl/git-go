package main

import (
	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/pathutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type flags struct {
	C pflag.Value // simpler version of git's -C: https://git-scm.com/docs/git#Documentation/git.txt--Cltpathgt

	env *env.Env
}

func newRootCmd(cwd string, e *env.Env) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "git-go",
		Short:         "git implementation in pure Go",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cfg := &flags{
		env: e,
	}
	cfg.C = pathutil.NewDirPathFlagWithDefault(cwd)
	cmd.PersistentFlags().VarS(cfg.C, "C", "C", "Run as if git was started in the provided path instead of the current working directory.")

	// porcelain
	cmd.AddCommand(newInitCmd(cfg))

	// plumbing
	cmd.AddCommand(newCatFileCmd(cfg))
	cmd.AddCommand(newHashObjectCmd())

	return cmd
}
