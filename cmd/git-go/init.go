package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/spf13/cobra"
)

// initCmdFlags represents the flags accepted by the init command
//
// Reference: https://git-scm.com/docs/git-init#_options
type initCmdFlags struct {
	initialBranch  string
	separateGitDir string
	quiet          bool
}

func newInitCmd(cfg *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init a new git repository",
	}

	flags := initCmdFlags{}
	cmd.Flags().StringVarP(&flags.initialBranch, "initial-branch", "b", "", "Use the specified name for the initial branch in the newly created repository. If not specified, fall back to the default name (currently master, but this is subject to change in the future; the name can be customized via the init.defaultBranch configuration variable).")
	cmd.Flags().BoolVarP(&flags.quiet, "quiet", "q", false, "Only print error and warning messages; all other output will be suppressed.")
	cmd.Flags().StringVar(&flags.separateGitDir, "separate-git-dir", "", "Instead of initializing the repository as a directory to either $GIT_DIR or ./.git/, create a text file there containing the path to the actual repository. This file acts as filesystem-agnostic Git symbolic link to the repository.\n\nIf this is reinitialization, the repository will be moved to the specified path.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return initCmd(cmd.OutOrStdout(), cfg, flags)
	}

	return cmd
}

func initCmd(out io.Writer, cfg *globalFlags, flags initCmdFlags) error {
	// validate conflicting options
	gitDir := cfg.GitDir
	if flags.separateGitDir != "" {
		if cfg.Bare {
			return errors.New("--separate-git-dir and --bare are mutually exclusive")
		}

		if cfg.GitDir != "" || cfg.env.Get("GIT_DIR") != "" {
			return errors.New("fatal: --separate-git-dir incompatible with bare repository")
		}
		gitDir = flags.separateGitDir
	}

	p, err := config.LoadConfig(cfg.env, config.LoadConfigOptions{
		WorkingDirectory: cfg.C.String(),
		GitDirPath:       gitDir,
		WorkTreePath:     cfg.WorkTree,
		IsBare:           cfg.Bare,
		SkipGitDirLookUp: true,
	})
	if err != nil {
		return fmt.Errorf("could not create param: %w", err)
	}

	// Let's check if the repo already exists by cheking is a HEAD is
	// in there
	newRepo := true
	_, err = os.Stat(filepath.Join(ginternals.DotGitPath(p), ginternals.Head))
	if err == nil {
		newRepo = false
	}

	r, err := git.InitRepositoryWithParams(p, git.InitOptions{
		IsBare:            cfg.Bare,
		InitialBranchName: flags.initialBranch,
		Symlink:           flags.separateGitDir != "",
	})
	if err != nil {
		return err
	}

	if !flags.quiet {
		switch newRepo {
		case true:
			fmt.Fprintln(out, "Initialized empty Git repository in", ginternals.DotGitPath(r.Config))
		case false:
			fmt.Fprintln(out, "Reinitialized existing Git repository in", ginternals.DotGitPath(r.Config))
		}
	}

	return r.Close()
}
