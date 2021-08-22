package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Nivl/git-go/ginternals/object"
	"github.com/spf13/cobra"
)

func newHashObjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash-object FILE",
		Short: "Compute object ID and optionally creates a blob from a file",
		Args:  cobra.ExactArgs(1),
	}

	typ := cmd.Flags().StringS("type", "t", "blob", "Specify the type")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return hashObjectCmd(cmd.OutOrStdout(), args[0], *typ)
	}

	return cmd
}

func hashObjectCmd(out io.Writer, filePath, typ string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read file content: %w", err)
	}

	var o *object.Object
	switch typ {
	case object.TypeBlob.String():
		o = object.New(object.TypeBlob, content)
	case object.TypeCommit.String():
		o = object.New(object.TypeCommit, content)
		_, err = o.AsCommit()
		if err != nil {
			return fmt.Errorf("invalid commit file: %w", err)
		}
	case object.TypeTree.String():
		o = object.New(object.TypeTree, content)
		_, err = o.AsTree()
		if err != nil {
			return fmt.Errorf("invalid tree file: %w", err)
		}
	case object.TypeTag.String():
		fallthrough
	default:
		return fmt.Errorf("unsupported object type %s", typ)
	}

	_, err = o.Compress()
	if err != nil {
		return fmt.Errorf("could not compress file: %w", err)
	}

	fmt.Fprintln(out, o.ID().String())
	return nil
}
