// Package object contains methods and objects to work with git objects
package object

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"strconv"

	"errors"

	"github.com/Nivl/git-go/internal/readutil"
	"github.com/Nivl/git-go/plumbing"
	"golang.org/x/xerrors"
)

var (
	// ErrObjectUnknown represents an error thrown when encoutering an
	// unknown object
	ErrObjectUnknown = errors.New("invalid object type")

	// ErrTreeInvalid represents an error thrown when parsing an invalid
	// tree object
	ErrTreeInvalid = errors.New("invalid tree")

	// ErrCommitInvalid represents an error thrown when parsing an invalid
	// commit object
	ErrCommitInvalid = errors.New("invalid commit")
)

// Type represents the type of an object as stored in a packfile
type Type int8

// List of all the possible object types
const (
	TypeCommit Type = 1
	TypeTree   Type = 2
	TypeBlob   Type = 3
	TypeTag    Type = 4
	// 5 is reserved for future use
	ObjectDeltaOFS Type = 6
	ObjectDeltaRef Type = 7
)

func (t Type) String() string {
	switch t {
	case TypeCommit:
		return "commit"
	case TypeTree:
		return "tree"
	case TypeBlob:
		return "blob"
	case TypeTag:
		return "tag"
	case ObjectDeltaOFS:
		return "osf-delta"
	case ObjectDeltaRef:
		return "ref-delta"
	default:
		panic(fmt.Sprintf("unknown object type %d", t))
	}
}

// IsValid check id the object type is an existing type
func (t Type) IsValid() bool {
	switch t {
	case TypeCommit,
		TypeTree,
		TypeBlob,
		TypeTag,
		ObjectDeltaOFS,
		ObjectDeltaRef:
		return true
	default:
		return false
	}
}

// NewTypeFromString returns an Type from its string
// representation
func NewTypeFromString(t string) (Type, error) {
	switch t {
	case "commit":
		return TypeCommit, nil
	case "tree":
		return TypeTree, nil
	case "blob":
		return TypeBlob, nil
	case "tag":
		return TypeTag, nil
	default:
		return 0, ErrObjectUnknown
	}
}

// Object represents a git object. An object can be of multiple types
// but they all share similarities (same storage system, same header,
// etc.).
// Object are stored in .git/objects, and may be stored in a packfile
// (kind of an optimized git database) located in .git/objects/packs
// https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
type Object struct {
	ID      plumbing.Oid
	typ     Type
	content []byte
}

// New creates a new git object of the given type
// The Object ID won't be calculated until Compress() is called
func New(typ Type, content []byte) *Object {
	return &Object{
		ID:      plumbing.NullOid,
		typ:     typ,
		content: content,
	}
}

// NewWithID creates a new git object of the given type with the given ID
func NewWithID(id plumbing.Oid, typ Type, content []byte) *Object {
	return &Object{
		ID:      id,
		typ:     typ,
		content: content,
	}
}

// Size returns the size of the object
func (o *Object) Size() int {
	return len(o.content)
}

// Type returns the Type for this object
func (o *Object) Type() Type {
	return o.typ
}

// Bytes returns the object's contents
func (o *Object) Bytes() []byte {
	return o.content
}

// Compress return the object zlib compressed, alongside its oid.
// The format of the compressed data is:
// [type] [size][NULL][content]
// The type in ascii, followed by a space, followed by the size in ascii,
// followed by a null character (0), followed by the object data
// maybe we can move some code around
func (o *Object) Compress() (oid plumbing.Oid, data []byte, err error) {
	// Quick reminder that the Write* methods on bytes.Buffer never fails,
	// the error returned is always nil
	w := new(bytes.Buffer)

	// Write the type
	w.WriteString(o.Type().String())
	// add the space
	w.WriteRune(' ')
	// write the size
	w.WriteString(strconv.Itoa(o.Size()))
	// Write the NULL char
	w.WriteByte(0)
	// Write the content
	w.Write(o.Bytes())

	// get the SHA of the file
	fileContent := w.Bytes()
	o.ID = plumbing.NewOid(fileContent)

	compressedContent := new(bytes.Buffer)
	zw := zlib.NewWriter(compressedContent)
	defer func() {
		closeErr := zw.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if _, err = zw.Write(fileContent); err != nil {
		return plumbing.NullOid, nil, xerrors.Errorf("could not zlib the object: %w", err)
	}
	if err = zw.Close(); err != nil {
		return plumbing.NullOid, nil, xerrors.Errorf("could not close the compressor: %w", err)
	}
	return o.ID, compressedContent.Bytes(), nil
}

// AsBlob parses the object as Blob
func (o *Object) AsBlob() *Blob {
	return NewBlob(o.ID, o.content)
}

// AsTree parses the object as Tree
//
// A commit has following format:
//
// {octal_mode} {path_name}\0{encoded_sha}
//
// Note:
// - a Tree may have multiple entries
func (o *Object) AsTree() (*Tree, error) {
	entries := []*TreeEntry{}

	objData := o.Bytes()
	offset := 0
	var err error
	for {
		entry := &TreeEntry{}
		data := readutil.ReadTo(objData[offset:], ' ')
		if len(data) == 0 {
			return nil, xerrors.Errorf("could not retrieve the mode: %w", ErrTreeInvalid)
		}
		offset += len(data) + 1 // +1 for the space
		entry.Mode = string(data)

		data = readutil.ReadTo(objData[offset:], 0)
		if len(data) == 0 {
			return nil, xerrors.Errorf("could not retrieve the path: %w", ErrTreeInvalid)
		}
		offset += len(data) + 1 // +1 for the \0
		entry.Path = string(data)

		if offset+20 > len(objData) {
			return nil, xerrors.Errorf("not enough space to retrieve the ID: %w", ErrTreeInvalid)
		}
		entry.ID, err = plumbing.NewOidFromHex(objData[offset : offset+20])
		if err != nil {
			return nil, xerrors.Errorf("%s: %w", ErrTreeInvalid.Error(), err)
		}
		offset += 20

		entries = append(entries, entry)
		if len(objData) == offset {
			break
		}
	}

	return NewTree(o.ID, entries), nil
}

// AsCommit parses the object as Commit
//
// A commit has following format:
//
// tree {sha}
// parent {sha}
// author {author_name} <{author_email}> {author_date_seconds} {author_date_timezone}
// committer {committer_name} <{committer_email}> {committer_date_seconds} {committer_date_timezone}
// gpgsig -----BEGIN PGP SIGNATURE-----
// {gpg key over multiple lines}
//  -----END PGP SIGNATURE-----
// {a blank line}
// {commit message}
//
// Note:
// - A commit can have 0, 1, or many parents lines
//   The very first commit of a repo has no parents
//   A regular commit as 1 parent
//   A merge commit has 2 or more parents
// - The gpgsig is optional
func (o *Object) AsCommit() (*Commit, error) {
	if o.typ != TypeCommit {
		return nil, xerrors.Errorf("type %s is not a commit", o.typ)
	}
	ci := &Commit{ID: o.ID}
	offset := 0
	objData := o.Bytes()
	for {
		line := readutil.ReadTo(objData[offset:], '\n')
		offset += len(line) + 1 // +1 to count the \n

		// If we didn't find anything then something is wrong
		if len(line) == 0 && offset == 1 {
			return nil, xerrors.Errorf("could not find commit first line: %w", ErrCommitInvalid)
		}

		// if we got an empty line, it means everything from now to the end
		// will be the commit message
		if len(line) == 0 {
			ci.Message = string(objData[offset:])
			break
		}

		// Otherwise we're getting a key/value pair, separated by a space
		kv := bytes.SplitN(line, []byte{' '}, 2)
		switch string(kv[0]) {
		case "tree":
			oid, err := plumbing.NewOidFromChars(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse tree id %#v: %w", kv[1], err)
			}
			ci.TreeID = oid
		case "parent":
			oid, err := plumbing.NewOidFromChars(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse parent id %#v: %w", kv[1], err)
			}
			ci.ParentIDs = append(ci.ParentIDs, oid)
		case "author":
			sig, err := NewSignatureFromBytes(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse signature [%s]: %w", string(kv[1]), err)
			}
			ci.Author = sig
		case "committer":
			sig, err := NewSignatureFromBytes(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse signature [%s]: %w", string(kv[1]), err)
			}
			ci.Committer = sig
		case "gpgsig":
			begin := string(kv[1]) + "\n"
			end := "-----END PGP SIGNATURE-----\n"
			i := bytes.Index(objData[offset:], []byte(end))
			ci.GPGSig = begin + string(objData[offset:offset+i]) + end
			offset += len(end) + i
		}
	}

	return ci, nil
}
