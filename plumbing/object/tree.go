package object

import (
	"bytes"
	"os"
	"strconv"

	"github.com/Nivl/git-go/plumbing"
)

// Tree represents a git tree object
type Tree struct {
	ID      plumbing.Oid
	Entries []*TreeEntry
}

// TreeEntry represents an entry inside a git tree
type TreeEntry struct {
	Mode os.FileMode
	ID   plumbing.Oid
	Path string
}

// NewTree returns a new tree with the given entries
func NewTree(entries []*TreeEntry) *Tree {
	return &Tree{
		Entries: entries,
	}
}

// NewTreeWithID returns a new tree
func NewTreeWithID(id plumbing.Oid, entries []*TreeEntry) *Tree {
	return &Tree{
		ID:      id,
		Entries: entries,
	}
}

// ToObject returns an Object representing the tree
func (t *Tree) ToObject() (*Object, error) {
	// Quick reminder that the Write* methods on bytes.Buffer never fails,
	// the error returned is always nil
	buf := new(bytes.Buffer)

	// The format of an tree entry is:
	// {octal_mode} {path_name}\0{encoded_sha}
	// A tree object is only composed of a bunch of entries back to back
	for _, e := range t.Entries {
		// Write the mode
		buf.WriteString(strconv.FormatInt(int64(e.Mode), 8))
		// add space
		buf.WriteByte(' ')
		// add the path
		buf.WriteString(e.Path)
		// Write the NULL char
		buf.WriteByte(0)
		// Finish with the encoded oid
		buf.Write(e.ID.Bytes())
	}

	if t.ID != plumbing.NullOid {
		return NewWithID(t.ID, TypeTree, buf.Bytes()), nil
	}
	return New(TypeTree, buf.Bytes()), nil
}
