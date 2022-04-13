package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/errutil"
	"github.com/spf13/cobra"
)

// switchCmdFlags represents the flags accepted by the switch command
//
// Reference: https://git-scm.com/docs/git-switch#_options
type switchCmdFlags struct {
	createBranch      string
	forceCreateBranch string
	orphan            string
	quiet             bool
	detach            bool
}

func newSwitchCmd(cfg *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch [branch|start-point]",
		Short: "Switch branches",
		Long:  "Switch to a specified branch. The working tree and the index are updated to match the branch. All new commits will be added to the tip of this branch.\n\nOptionally a new branch could be created with either -c, -C, automatically from a remote branch of same name (see --guess), or detach the working tree from any branch with --detach, along with switching.\n\nSwitching branches does not require a clean index and working tree (i.e. no differences compared to HEAD). The operation is aborted however if the operation leads to loss of local changes, unless told otherwise with --discard-changes or --merge.",
		Args:  cobra.MaximumNArgs(1),
	}

	flags := switchCmdFlags{}
	cmd.Flags().StringVarP(&flags.createBranch, "create", "c", "", "Create a new branch named <new-branch> starting at <start-point> before switching to the branch.")
	// We can't use C because it's already used by the root command
	cmd.Flags().StringVar(&flags.forceCreateBranch, "force-create", "", "Similar to --create except that if <new-branch> already exists, it will be reset to <start-point>.")
	cmd.Flags().StringVar(&flags.orphan, "orphan", "", "Create a new orphan branch, named <new-branch>. All tracked files are removed.")
	cmd.Flags().BoolVarP(&flags.detach, "detach", "d", false, "Switch to a commit for inspection and discardable experiments.")
	cmd.Flags().BoolVarP(&flags.quiet, "quiet", "q", false, "Quiet, suppress feedback messages.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		directory := ""
		if len(args) > 0 {
			directory = args[0]
		}
		return switchCmd(cmd.OutOrStdout(), cfg, flags, directory)
	}

	return cmd
}

// TODO(melvin):
// --ignore-other-worktrees
// -t, --track [direct|inherit]
// --no-track
// --recurse-submodules, --no-recurse-submodules
// --progress, --no-progress
// -m, --merge
// --conflict=<style>
// -f, --force, --discard-changes
// --guess, --no-guess
func switchCmd(out io.Writer, cfg *globalFlags, flags switchCmdFlags, starterPointOrBranch string) (err error) {
	r, err := loadRepository(cfg)
	if err != nil {
		return fmt.Errorf("could not create param: %w", err)
	}
	defer errutil.Close(r, &err)

	// validate conflicting options
	switch {
	case flags.detach:
		if flags.createBranch != "" || flags.forceCreateBranch != "" || flags.orphan != "" {
			return errors.New("'--detach' cannot be used with '-c/-C/--orphan'")
		}

		refName := "HEAD"
		if starterPointOrBranch != "" {
			refName = ginternals.LocalBranchFullName(starterPointOrBranch)
		}

		ref, err := r.Reference(refName)
		if err != nil && !errors.Is(err, ginternals.ErrRefNotFound) {
			return fmt.Errorf("couldn't get '%s': %w", flags.orphan, err)
		}

		isRef := err == nil
		oid := ginternals.NullOid
		switch isRef {
		case true:
			oid = ref.Target()
		case false: // We either have a commit, or something invalid
			oid, err = ginternals.NewOidFromStr(starterPointOrBranch)
			if err != nil {
				return fmt.Errorf("invalid branch or sha '%s'", flags.orphan)
			}
		}

		c, err := r.Commit(oid)
		if err != nil {
			return fmt.Errorf("couldn't get commit '%s': %w", oid.String(), err)
		}

		_, err = r.NewReference(ginternals.Head, oid)
		if err != nil {
			return fmt.Errorf("couldn't update HEAD: %w", err)
		}

		// TODO(melvin): clean the index

		fprintf(flags.quiet, out, "HEAD is now at %s %s", oid.String(), c.Message())
	case flags.orphan != "":
		if flags.createBranch != "" || flags.forceCreateBranch != "" {
			return errors.New("options '-c', and '--orphan' cannot be used together")
		}
		if starterPointOrBranch != "" {
			return errors.New("'--orphan' cannot take <start-point>")
		}

		// Let's make sure a branch with the same name doesn't exist
		_, err = r.Reference(ginternals.LocalBranchFullName(flags.orphan))
		if !errors.Is(err, ginternals.ErrRefNotFound) {
			if err == nil {
				return fmt.Errorf("a branch named '%s' already exists", flags.orphan)
			}
			return fmt.Errorf("couldn't get '%s': %w", flags.orphan, err)
		}

		// TODO(melvin): remove all tracked files before switching
		// Let's set the current branch as current
		// The ref will be dangling since it doesn't have commits
		_, err = r.NewSymbolicReference(ginternals.Head, ginternals.LocalBranchFullName(flags.orphan))
		if err != nil {
			return fmt.Errorf("couldn't update HEAD: %w", err)
		}

		// TODO(melvin): clean the index

		fprintf(flags.quiet, out, "Switched to a new branch '%s'\n", flags.orphan)
	default:
		if starterPointOrBranch == "" {
			return errors.New("missing branch or commit argument")
		}
		// TODO(melvin): handle dangling refs

		// We make sure we're not already on the branch
		head, err := r.Reference(ginternals.Head)
		if err != nil {
			return fmt.Errorf("couldn't load %s: %w", ginternals.Head, err)
		}
		if head.SymbolicTarget() == ginternals.LocalBranchFullName(starterPointOrBranch) {
			fprintf(flags.quiet, out, "Already on '%s'\n", starterPointOrBranch)
			// TODO(melvin): check if branch is up-to-date with remote
			return nil
		}

		// We make sure the target branch already exists
		_, err = r.Reference(ginternals.LocalBranchFullName(starterPointOrBranch))
		if err != nil {
			return fmt.Errorf("couldn't load %s: %w", starterPointOrBranch, err)
		}

		// TODO(melvin): abort if there are conflicts between the wt/index
		// of the branches

		// Set the current branch
		_, err = r.NewSymbolicReference(ginternals.Head, ginternals.LocalBranchFullName(starterPointOrBranch))
		if err != nil {
			return fmt.Errorf("couldn't update HEAD: %w", err)
		}

		// TODO(melvin): clean the index

		// TODO(melvin): check if branch is up-to-date with remote
		fprintf(flags.quiet, out, "Switched to branch '%s'\n", starterPointOrBranch)
	}

	return nil
}
