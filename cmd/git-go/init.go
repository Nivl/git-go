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
		Use:   "init [directory]",
		Short: "init a new git repository",
		Long:  "This command creates an empty Git repository - basically a .git directory with subdirectories for objects, refs/heads, refs/tags, and template files. An initial branch without any commits will be created (see the --initial-branch option below for its name).\n\nIf the $GIT_DIR environment variable is set then it specifies a path to use instead of ./.git for the base of the repository.\n\nIf the object storage directory is specified via the $GIT_OBJECT_DIRECTORY environment variable then the sha1 directories are created underneath - otherwise the default $GIT_DIR/objects directory is used.\n\nRunning git init in an existing repository is safe. It will not overwrite things that are already there. The primary reason for rerunning git init is to pick up newly added templates (or to move the repository to another place if --separate-git-dir is given).",
		Args:  cobra.MaximumNArgs(1),
	}

	flags := initCmdFlags{}
	cmd.Flags().StringVarP(&flags.initialBranch, "initial-branch", "b", "", "Use the specified name for the initial branch in the newly created repository. If not specified, fall back to the default name (currently master, but this is subject to change in the future; the name can be customized via the init.defaultBranch configuration variable).")
	cmd.Flags().BoolVarP(&flags.quiet, "quiet", "q", false, "Only print error and warning messages; all other output will be suppressed.")
	cmd.Flags().StringVar(&flags.separateGitDir, "separate-git-dir", "", "Instead of initializing the repository as a directory to either $GIT_DIR or ./.git/, create a text file there containing the path to the actual repository. This file acts as filesystem-agnostic Git symbolic link to the repository.\n\nIf this is reinitialization, the repository will be moved to the specified path.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		directory := ""
		if len(args) > 0 {
			directory = args[0]
		}
		return initCmd(cmd.OutOrStdout(), cfg, flags, directory)
	}

	return cmd
}

func initCmd(out io.Writer, cfg *globalFlags, flags initCmdFlags, optionalDirectory string) error {
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

	workingDirectory := cfg.C.String()
	if optionalDirectory != "" {
		workingDirectory = optionalDirectory
	}

	p, err := config.LoadConfig(cfg.env, config.LoadConfigOptions{
		WorkingDirectory: workingDirectory,
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

	switch newRepo {
	case true:
		fprintln(flags.quiet, out, "Initialized empty Git repository in", ginternals.DotGitPath(r.Config))
	case false:
		fprintln(flags.quiet, out, "Reinitialized existing Git repository in", ginternals.DotGitPath(r.Config))
	}

	return r.Close()
}
