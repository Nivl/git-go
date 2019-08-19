package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/goabstract/git"
	"github.com/pkg/errors"

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
		Use:   "agit",
		Short: "abstract's git implementation in pure Go",
	}

	cmd.AddCommand(newInitCmd())
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
		Use:   "cat-file TYPE OBJECT",
		Short: "Provide content or type and size information for repository objects",
		Args:  cobra.ExactArgs(2),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return catFileCmd(args[0], args[1])
	}

	return cmd
}

func catFileCmd(typ, sha string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	r, err := git.LoadRepository(pwd)
	if err != nil {
		return err
	}

	oid, err := git.NewOidFromStr(sha)
	if err != nil {
		return errors.Wrap(err, "failed parsing sha")
	}

	o, err := r.GetObject(oid)
	if err != nil {
		return err
	}

	fmt.Print(string(o.Bytes()))
	return nil
}

func newHashObjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash-object FILE",
		Short: "Compute object ID and optionally creates a blob from a file",
		Args:  cobra.ExactArgs(1),
	}

	typ := cmd.Flags().StringP("type", "t", "blob", "Specify the type")
	write := cmd.Flags().BoolP("write", "w", false, "Actually write the object into the object database")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return hashObjectCmd(args[0], *typ, *write)
	}

	return cmd
}

func hashObjectCmd(filePath, typ string, write bool) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var o *git.Object
	switch typ {
	case "blob":
		o = git.NewObject(git.ObjectTypeBlob, content)
	case "commit":
		o = git.NewObject(git.ObjectTypeCommit, content)
	case "tree":
		o = git.NewObject(git.ObjectTypeTree, content)
	case "tag":
		fallthrough
	default:
		return errors.Errorf("unsupported object type %s", typ)
	}

	oid, _, err := o.Compress()
	if err != nil {
		return err
	}

	fmt.Println(oid.String())
	return nil
}
