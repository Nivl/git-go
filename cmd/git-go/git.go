package main

import (
	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/pathutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type globalFlags struct {
	env *env.Env

	C        pflag.Value // simpler version of git's -C: https://git-scm.com/docs/git#Documentation/git.txt--Cltpathgt
	WorkTree string
	GitDir   string
	Bare     bool
}

func newRootCmd(cwd string, e *env.Env) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "git-go",
		Short:         "git implementation in pure Go",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cfg := &globalFlags{
		env: e,
	}
	cfg.C = pathutil.NewDirPathFlagWithDefault(cwd)
	cmd.PersistentFlags().VarS(cfg.C, "C", "C", "Run as if git was started in the provided path instead of the current working directory.")
	cmd.PersistentFlags().BoolVar(&cfg.Bare, "bare", false, "Treat the repository as a bare repository")
	cmd.PersistentFlags().StringVar(&cfg.GitDir, "git-dir", "", "Set the path to the repository")
	cmd.PersistentFlags().StringVar(&cfg.WorkTree, "work-tree", "", "Set the path to the root of the working tree")

	// porcelain
	cmd.AddCommand(newInitCmd(cfg))
	cmd.AddCommand(newSwitchCmd(cfg))

	// plumbing
	cmd.AddCommand(newCatFileCmd(cfg))
	cmd.AddCommand(newHashObjectCmd())

	return cmd
}
