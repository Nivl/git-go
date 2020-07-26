package object

import (
	"bytes"
	"strconv"

	"github.com/Nivl/git-go/plumbing"
)

// TreeObjectMode represents the mode of an object inside a tree
// Non-standard modes (like 0o100664) are not supported
type TreeObjectMode int32

const (
	// ModeFile represents the mode to use for a regular file
	ModeFile TreeObjectMode = 0o100644
	// ModeExecutable represents the mode to use for a executable file
	ModeExecutable TreeObjectMode = 0o100755
	// ModeDirectory represents the mode to use for a directory
	ModeDirectory TreeObjectMode = 0o040000
	// ModeSymLink represents the mode to use for a symbolic link
	ModeSymLink TreeObjectMode = 0o120000
	// ModeGitLink represents the mode to use for a gitlink (submodule)
	ModeGitLink TreeObjectMode = 0o160000
)

// IsValid returns whether the mode is a supported mode or not
func (m TreeObjectMode) IsValid() bool {
	// we use a switch because any missing value will be detected
	// by our linter
	switch m {
	case ModeFile, ModeExecutable, ModeDirectory, ModeSymLink, ModeGitLink:
		return true
	default:
		return false
	}
}

// Tree represents a git tree object
type Tree struct {
	id plumbing.Oid
	// we don't use pointers to make sure entries are immutable
	entries []TreeEntry
}

// TreeEntry represents an entry inside a git tree
type TreeEntry struct {
	Mode TreeObjectMode
	ID   plumbing.Oid
	Path string
}

// NewTree returns a new tree with the given entries
func NewTree(entries []TreeEntry) *Tree {
	return &Tree{
		entries: entries,
	}
}

// NewTreeWithID returns a new tree
func NewTreeWithID(id plumbing.Oid, entries []TreeEntry) *Tree {
	return &Tree{
		id:      id,
		entries: entries,
	}
}

// Entries returns a copy of tree entries
func (t *Tree) Entries() []TreeEntry {
	out := make([]TreeEntry, len(t.entries))
	copy(out, t.entries)
	return out
}

// ID returns the object's ID
// plumbing.NullOid is returned if the object doesn't have
// an ID yet
func (t *Tree) ID() plumbing.Oid {
	return t.id
}

// ToObject returns an Object representing the tree
func (t *Tree) ToObject() *Object {
	// Quick reminder that the Write* methods on bytes.Buffer never fails,
	// the error returned is always nil
	buf := new(bytes.Buffer)

	// The format of an tree entry is:
	// {octal_mode} {path_name}\0{encoded_sha}
	// A tree object is only composed of a bunch of entries back to back
	for _, e := range t.entries {
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

	if t.id != plumbing.NullOid {
		return NewWithID(t.id, TypeTree, buf.Bytes())
	}
	return New(TypeTree, buf.Bytes())
}
