package object

import (
	"bytes"
	"strconv"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/readutil"
	"golang.org/x/xerrors"
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

// ObjectType returns the object type associated to a mode
func (m TreeObjectMode) ObjectType() Type {
	switch m {
	case ModeDirectory:
		return TypeTree
	case ModeGitLink:
		return TypeCommit
	case ModeExecutable, ModeFile, ModeSymLink:
		return TypeBlob
	default:
		// We treat anything unexpected as blob
		return TypeBlob
	}
}

// Tree represents a git tree object
type Tree struct {
	rawObject *Object
	// we don't use pointers to make sure entries are immutable
	entries []TreeEntry
}

// TreeEntry represents an entry inside a git tree
type TreeEntry struct {
	Path string
	ID   ginternals.Oid
	Mode TreeObjectMode
}

// NewTree returns a new tree with the given entries
func NewTree(entries []TreeEntry) *Tree {
	t := &Tree{
		entries: entries,
	}
	t.rawObject = t.ToObject()
	return t
}

// NewTreeFromObject returns a new tree from an object
//
// A tree has following format:
//
// {octal_mode} {path_name}\0{encoded_sha}
//
// Note:
// - a Tree may have multiple entries
func NewTreeFromObject(o *Object) (*Tree, error) {
	if o.Type() != TypeTree {
		return nil, xerrors.Errorf("type %s is not a tree: %w", o.typ, ErrObjectInvalid)
	}

	entries := []TreeEntry{}

	objData := o.Bytes()
	if len(objData) > 0 {
		offset := 0
		// the variable i is only use for logs and error messages, not for
		// actual processing
		for i := 1; ; i++ {
			entry := TreeEntry{}
			data := readutil.ReadTo(objData[offset:], ' ')
			if len(data) == 0 {
				return nil, xerrors.Errorf("could not retrieve the mode of entry %d: %w", i, ErrTreeInvalid)
			}
			offset += len(data) + 1 // +1 for the space
			mode, err := strconv.ParseInt(string(data), 8, 32)
			if err != nil {
				return nil, xerrors.Errorf("could not parse mode of entry %d: %s: %w", i, err.Error(), ErrTreeInvalid)
			}
			entry.Mode = TreeObjectMode(mode)

			data = readutil.ReadTo(objData[offset:], 0)
			if len(data) == 0 {
				return nil, xerrors.Errorf("could not retrieve the path of entry %d: %w", i, ErrTreeInvalid)
			}
			offset += len(data) + 1 // +1 for the \0
			entry.Path = string(data)

			if offset+20 > len(objData) {
				return nil, xerrors.Errorf("not enough space to retrieve the ID of entry %d: %w", i, ErrTreeInvalid)
			}
			entry.ID, err = ginternals.NewOidFromHex(objData[offset : offset+20])
			if err != nil {
				// should never fail since any value is valid as long as it
				// is 20 chars
				return nil, xerrors.Errorf("invalid SHA for entry %d (%s): %w", i, err.Error(), ErrTreeInvalid)
			}
			offset += 20

			entries = append(entries, entry)
			if len(objData) == offset {
				break
			}
		}
	}
	return &Tree{
		rawObject: o,
		entries:   entries,
	}, nil
}

// Entries returns a copy of tree entries
func (t *Tree) Entries() []TreeEntry {
	out := make([]TreeEntry, len(t.entries))
	copy(out, t.entries)
	return out
}

// ID returns the object's ID
// ginternals.NullOid is returned if the object doesn't have
// an ID yet
func (t *Tree) ID() ginternals.Oid {
	return t.rawObject.ID()
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

	return New(TypeTree, buf.Bytes())
}
