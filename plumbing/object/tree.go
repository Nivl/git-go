package object

import (
	"os"

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

// NewTree returns a new tree from an object
func NewTree(id plumbing.Oid, entries []*TreeEntry) *Tree {
	return &Tree{
		ID:      id,
		Entries: entries,
	}
}
