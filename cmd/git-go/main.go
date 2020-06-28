package main

import (
	"fmt"
	"os"

	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/plumbing"
	"golang.org/x/xerrors"

	"github.com/spf13/cobra"
)

func main() {
	root := newRootCmd()
	err := root.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "git-go",
		Short:         "git implementation in pure Go",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// porcelain
	cmd.AddCommand(newInitCmd())

	// plumbing
	cmd.AddCommand(newCatFileCmd())
	cmd.AddCommand(newHashObjectCmd())

	return cmd
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init a new git repository",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return initCmd()
	}

	return cmd
}

func initCmd() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	_, err = git.InitRepository(pwd)
	return err
}

func newCatFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cat-file OBJECT",
		Short: "Provide content or type and size information for repository objects",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return catFileCmd(args[0])
	}

	return cmd
}

func catFileCmd(sha string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	r, err := git.LoadRepository(pwd)
	if err != nil {
		return err
	}

	oid, err := plumbing.NewOidFromStr(sha)
	if err != nil {
		return xerrors.Errorf("failed parsing sha: %w", err)
	}

	o, err := r.GetObject(oid)
	if err != nil {
		return err
	}

	fmt.Print(string(o.Bytes()))
	return nil
}
