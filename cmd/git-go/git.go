package main

import (
	"strings"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/pathutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type config struct {
	C                pflag.Value // simpler version of git's -C: https://git-scm.com/docs/git#Documentation/git.txt--Cltpathgt
	GitDir           string      // Loaded from env $GIT_DIR
	GitObjectDir     string      // Loaded from env $GIT_OBJECT_DIRECTORY
	GitConfig        string      // Loaded from env $GIT_CONFIG
	SkipSystemConfig bool        // Loaded from env $GIT_CONFIG_NOSYSTEM
}

func newRootCmd(cwd string, e *env.Env) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "git-go",
		Short:         "git implementation in pure Go",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	SkipSystemConfig := false
	switch strings.ToLower(e.Get("GIT_CONFIG_NOSYSTEM")) {
	case "yes", "1", "true":
		SkipSystemConfig = true
	}
	cfg := &config{
		GitDir:           e.Get("GIT_DIR"),
		GitObjectDir:     e.Get("GIT_OBJECT_DIRECTORY"),
		SkipSystemConfig: SkipSystemConfig,
		GitConfig:        e.Get("GIT_CONFIG"),
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
