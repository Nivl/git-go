package git

import (
	"fmt"
	"sort"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
)

// TreeBuilder is used to build trees
type TreeBuilder struct {
	Backend *backend.Backend
	entries map[string]object.TreeEntry
}

// NewTreeBuilder create a new empty tree builder
func (r *Repository) NewTreeBuilder() *TreeBuilder {
	return &TreeBuilder{
		Backend: r.dotGit,
	}
}

// NewTreeBuilderFromTree create a new tree builder containing the
// entries of another tree
func (r *Repository) NewTreeBuilderFromTree(t *object.Tree) *TreeBuilder {
	entries := map[string]object.TreeEntry{}
	for _, e := range t.Entries() {
		entries[e.Path] = e
	}

	return &TreeBuilder{
		Backend: r.dotGit,
		entries: entries,
	}
}

// Insert inserts a new object in a tree
func (tb *TreeBuilder) Insert(path string, oid ginternals.Oid, mode object.TreeObjectMode) error {
	if !mode.IsValid() {
		//nolint:goerr113 // no need to wrap the error, this would only be caused by a bug in the codebase
		return fmt.Errorf("invalid mode %o", mode)
	}

	o, err := tb.Backend.Object(oid)
	if err != nil {
		return fmt.Errorf("cannot verify object: %w", err)
	}

	// TODO(melvin):
	// 2. gitlink?
	if o.Type() != object.TypeBlob && o.Type() != object.TypeTree {
		return fmt.Errorf("unexpected object %s: %w", o.Type().String(), object.ErrObjectInvalid)
	}

	e := object.TreeEntry{
		Mode: mode,
		Path: path,
		ID:   oid,
	}

	if tb.entries == nil {
		tb.entries = map[string]object.TreeEntry{}
	}
	tb.entries[path] = e
	return nil
}

// Remove removes an object from tree
func (tb *TreeBuilder) Remove(path string) {
	if tb.entries == nil {
		return
	}
	delete(tb.entries, path)
}

// Write creates and persists a new Tree object
func (tb *TreeBuilder) Write() (*object.Tree, error) {
	// We need to order all our entries alphabetically
	// We're going to extract the paths of the map
	// and just loop over the keys instead of the entries
	paths := make([]string, 0, len(tb.entries))
	for p := range tb.entries {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	entries := make([]object.TreeEntry, 0, len(paths))
	for _, p := range paths {
		entries = append(entries, tb.entries[p])
	}

	t := object.NewTree(entries)
	o := t.ToObject()
	if _, err := tb.Backend.WriteObject(o); err != nil {
		return nil, fmt.Errorf("could not write the object to the odb: %w", err)
	}
	return o.AsTree()
}
