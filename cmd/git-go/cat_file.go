package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var errBadFile = errors.New("bad file")

func newCatFileCmd(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cat-file [TYPE] OBJECT",
		Short: "Provide content or type and size information for repository objects",
		Args:  cobra.RangeArgs(1, 2),
	}

	// TODO(melvin): Only use short flags
	// https://github.com/spf13/cobra/issues/679
	// https://github.com/spf13/pflag/pull/171
	// https://github.com/spf13/pflag/pull/256
	typeOnly := cmd.Flags().BoolP("t", "t", false, "Instead of the content, show the object type identified by <object>.")
	sizeOnly := cmd.Flags().BoolP("s", "s", false, "Instead of the content, show the object size identified by <object>.")
	prettyPrint := cmd.Flags().BoolP("p", "p", false, "Pretty-print the contents of <object> based on its type.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		p := catFileParams{
			typeOnly:    *typeOnly,
			sizeOnly:    *sizeOnly,
			prettyPrint: *prettyPrint,
			sha:         args[0],
		}
		if len(args) == 2 {
			p.typ = args[0]
			p.sha = args[1]
		}
		return catFileCmd(cmd.OutOrStdout(), cfg, p)
	}
	return cmd
}

type catFileParams struct {
	typeOnly    bool
	sizeOnly    bool
	prettyPrint bool
	sha         string
	typ         string
}

func catFileCmd(out io.Writer, cfg *config, p catFileParams) error {
	// Validate options
	if p.typ != "" && (p.typeOnly || p.sizeOnly || p.prettyPrint) {
		return errors.New("type not supported with options -t, -s, -p")
	}
	if p.typ == "" && !p.typeOnly && !p.sizeOnly && !p.prettyPrint {
		return errors.New("type and object required")
	}
	if p.typeOnly {
		switch {
		case p.sizeOnly:
			return errors.New("option -s not supported with option -t")
		case p.prettyPrint:
			return errors.New("option -p not supported with option -t")
		}
	}
	if p.sizeOnly && p.prettyPrint {
		return errors.New("option -p not supported with option -s")
	}

	// run the command
	r, err := loadRepository(cfg)
	if err != nil {
		return err
	}

	oid, err := plumbing.NewOidFromStr(p.sha)
	if err != nil {
		return xerrors.Errorf("failed parsing sha: %w", err)
	}

	o, err := r.GetObject(oid)
	if err != nil {
		return err
	}

	if p.typ != "" {
		_, err = object.NewTypeFromString(p.typ)
		if err != nil {
			return xerrors.Errorf("%s: %w", p.typ, err)
		}

		if o.Type().String() != p.typ {
			return xerrors.Errorf("%s: %w", p.sha, errBadFile)
		}
	}

	switch {
	case p.sizeOnly:
		fmt.Fprintln(out, strconv.Itoa(o.Size()))
	case p.typeOnly:
		fmt.Fprintln(out, o.Type().String())
	case p.prettyPrint:
		switch o.Type() {
		case object.TypeCommit:
			c, err := o.AsCommit()
			if err != nil {
				return xerrors.Errorf("could not get commit %w", err)
			}
			fmt.Fprintf(out, "tree %s\n", c.TreeID.String())
			for _, id := range c.ParentIDs {
				fmt.Fprintf(out, "parent %s\n", id.String())
			}
			fmt.Fprintf(out, "author %s\n", c.Author.String())
			fmt.Fprintf(out, "committer %s\n", c.Committer.String())
			if c.GPGSig != "" {
				fmt.Fprintf(out, "gpgsig %s \n", c.GPGSig)
			}
			fmt.Fprintln(out, "")
			fmt.Fprint(out, c.Message)
		case object.TypeTree:
			tree, err := o.AsTree()
			if err != nil {
				return xerrors.Errorf("could not get tree %w", err)
			}
			for _, e := range tree.Entries {
				// TODO(melvin): add the type
				fmt.Fprintf(out, "%06s %s\t%s\n", e.Mode, e.ID.String(), e.Path)
			}
		case object.TypeBlob:
			fmt.Fprint(out, string(o.Bytes()))
		case object.TypeTag, object.ObjectDeltaOFS, object.ObjectDeltaRef:
			fallthrough
		default:
			return xerrors.Errorf("pretty-print not supported for type %s", o.Type().String())
		}
	default:
		fmt.Fprint(out, string(o.Bytes()))
	}
	return nil
}
