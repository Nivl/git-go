package main

import (
	"os"

	"github.com/Nivl/git-go/internal/pathutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
)

type config struct {
	C pflag.Value // simpler version of git's -C: https://git-scm.com/docs/git#Documentation/git.txt--Cltpathgt
}

func newRootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:           "git-go",
		Short:         "git implementation in pure Go",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, xerrors.Errorf("couldn't get the current directory: %w", err)
	}

	cfg := &config{}
	cfg.C = pathutil.NewDirPathFlagWithDefault(pwd)
	cmd.PersistentFlags().VarP(cfg.C, "C", "C", "Run as if git was started in the provided path instead of the current working directory.")

	// porcelain
	cmd.AddCommand(newInitCmd(cfg))

	// plumbing
	cmd.AddCommand(newCatFileCmd(cfg))
	cmd.AddCommand(newHashObjectCmd())

	return cmd, nil
}
