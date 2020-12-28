package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/Nivl/git-go/internal/errutil"
	"github.com/Nivl/git-go/internal/gitpath"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
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
			objectName:  args[0],
		}
		if len(args) == 2 {
			p.typ = args[0]
			p.objectName = args[1]
		}
		return catFileCmd(cmd.OutOrStdout(), cfg, p)
	}
	return cmd
}

type catFileParams struct {
	typeOnly    bool
	sizeOnly    bool
	prettyPrint bool
	objectName  string
	typ         string
}

func catFileCmd(out io.Writer, cfg *config, p catFileParams) (err error) {
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
	defer errutil.Close(r, &err)

	oid, err := ginternals.NewOidFromStr(p.objectName)
	if err != nil {
		// If that failed it means we might have provided different name,
		// like a reference
		toTry := []string{
			// catches stuff like HEADS or refs/heads/master
			p.objectName,
			// catches heads/master
			gitpath.Ref(p.objectName),
			// catches local branch names
			gitpath.LocalBranch(p.objectName),
			// catches local tag names
			gitpath.LocalTag(p.objectName),
		}

		for _, refName := range toTry {
			ref, err := r.GetReference(refName)
			if err == nil {
				oid = ref.Target()
				break
			}

			// if the ref doesn't exist we test the the next one
			if !errors.Is(err, ginternals.ErrRefNotFound) {
				return xerrors.Errorf("could not check if ref %s exists: %w", refName, err)
			}
		}

		if oid.IsZero() {
			return xerrors.Errorf("not a valid object name %s", p.objectName)
		}
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
			return xerrors.Errorf("%s: %w", p.objectName, errBadFile)
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
			fmt.Fprintf(out, "tree %s\n", c.TreeID().String())
			for _, id := range c.ParentIDs() {
				fmt.Fprintf(out, "parent %s\n", id.String())
			}
			fmt.Fprintf(out, "author %s\n", c.Author().String())
			fmt.Fprintf(out, "committer %s\n", c.Committer().String())
			if c.GPGSig() != "" {
				fmt.Fprintf(out, "gpgsig %s \n", c.GPGSig())
			}
			fmt.Fprintln(out, "")
			fmt.Fprint(out, c.Message())
		case object.TypeTag:
			tag, err := o.AsTag()
			if err != nil {
				return xerrors.Errorf("could not get tag %w", err)
			}
			fmt.Fprintf(out, "object %s\n", tag.Target().String())
			fmt.Fprintf(out, "type %s\n", tag.Type().String())
			fmt.Fprintf(out, "tag %s\n", tag.Name())
			fmt.Fprintf(out, "tagger %s\n", tag.Tagger().String())
			if tag.GPGSig() != "" {
				fmt.Fprintf(out, "gpgsig %s \n", tag.GPGSig())
			}
			fmt.Fprintln(out, "")
			fmt.Fprint(out, tag.Message())
		case object.TypeTree:
			tree, err := o.AsTree()
			if err != nil {
				return xerrors.Errorf("could not get tree %w", err)
			}
			for _, e := range tree.Entries() {
				fmt.Fprintf(out, "%06o %s %s\t%s\n", e.Mode, e.Mode.ObjectType().String(), e.ID.String(), e.Path)
			}
		case object.TypeBlob:
			fmt.Fprint(out, string(o.Bytes()))
		case object.ObjectDeltaOFS, object.ObjectDeltaRef:
			fallthrough
		default:
			return xerrors.Errorf("pretty-print not supported for type %s", o.Type().String())
		}
	default:
		fmt.Fprint(out, string(o.Bytes()))
	}
	return nil
}
