// Package object contains methods and objects to work with git objects
package object

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/errutil"
	"github.com/Nivl/git-go/internal/readutil"
	"golang.org/x/xerrors"
)

var (
	// ErrObjectUnknown represents an error thrown when encoutering an
	// unknown object
	ErrObjectUnknown = errors.New("invalid object type")

	// ErrObjectInvalid represents an error thrown when an object contains
	// unexpected data or when the wrong object is provided to a method.
	// Ex. Inserting a ObjectDeltaOFS in a tree
	// Ex.2 Creating a tag using a commit with no ID (commit not persisted
	// 	to the odb)
	ErrObjectInvalid = errors.New("invalid object")

	// ErrTreeInvalid represents an error thrown when parsing an invalid
	// tree object
	ErrTreeInvalid = errors.New("invalid tree")

	// ErrCommitInvalid represents an error thrown when parsing an invalid
	// commit object
	ErrCommitInvalid = errors.New("invalid commit")

	// ErrTagInvalid represents an error thrown when parsing an invalid
	// tag object
	ErrTagInvalid = errors.New("invalid tag")
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
	id      ginternals.Oid
	typ     Type
	content []byte

	idProcessing sync.Once
}

// New creates a new git object of the given type
func New(typ Type, content []byte) *Object {
	o := &Object{
		typ:     typ,
		content: content,
	}
	o.id, _ = o.build()
	return o
}

// ID returns the ID of the object.
func (o *Object) ID() ginternals.Oid {
	o.idProcessing.Do(func() {
		o.id, _ = o.build()
	})
	return o.id
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

func (o *Object) build() (oid ginternals.Oid, data []byte) {
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
	data = w.Bytes()
	oid = ginternals.NewOidFromContent(data)
	return oid, data
}

// Compress return the object zlib compressed, alongside its oid.
// The format of the compressed data is:
// [type] [size][NULL][content]
// The type in ascii, followed by a space, followed by the size in ascii,
// followed by a null character (0), followed by the object data
// maybe we can move some code around
func (o *Object) Compress() (data []byte, err error) {
	// get the SHA of the file
	_, fileContent := o.build()

	compressedContent := new(bytes.Buffer)
	zw := zlib.NewWriter(compressedContent)
	defer errutil.Close(zw, &err)

	if _, err = zw.Write(fileContent); err != nil {
		return nil, xerrors.Errorf("could not zlib the object: %w", err)
	}
	return compressedContent.Bytes(), nil
}

// AsBlob parses the object as Blob
func (o *Object) AsBlob() *Blob {
	return NewBlob(o)
}

// AsTree parses the object as Tree
//
// A tree has following format:
//
// {octal_mode} {path_name}\0{encoded_sha}
//
// Note:
// - a Tree may have multiple entries
func (o *Object) AsTree() (*Tree, error) {
	entries := []TreeEntry{}

	objData := o.Bytes()
	offset := 0
	for i := 1; ; i++ {
		entry := TreeEntry{}
		data := readutil.ReadTo(objData[offset:], ' ')
		if len(data) == 0 {
			return nil, xerrors.Errorf("could not retrieve the mode of entry %d: %w", i, ErrTreeInvalid)
		}
		offset += len(data) + 1 // +1 for the space
		mode, err := strconv.ParseInt(string(data), 8, 32)
		if err != nil {
			return nil, xerrors.Errorf("could not parse mode of entry %d: %w", i, err)
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
			return nil, xerrors.Errorf("invalid SHA for entry %d (%s): %w", i, err.Error(), ErrTreeInvalid)
		}
		offset += 20

		entries = append(entries, entry)
		if len(objData) == offset {
			break
		}
	}

	return NewTreeWithID(o.ID(), entries), nil
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
	ci := &Commit{
		id:        o.ID(),
		rawObject: o,
	}
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
			ci.message = string(objData[offset:])
			break
		}

		// Otherwise we're getting a key/value pair, separated by a space
		kv := bytes.SplitN(line, []byte{' '}, 2)
		switch string(kv[0]) {
		case "tree":
			oid, err := ginternals.NewOidFromChars(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse tree id %#v: %w", kv[1], err)
			}
			ci.treeID = oid
		case "parent":
			oid, err := ginternals.NewOidFromChars(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse parent id %#v: %w", kv[1], err)
			}
			ci.parentIDs = append(ci.parentIDs, oid)
		case "author":
			sig, err := NewSignatureFromBytes(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse signature [%s]: %w", string(kv[1]), err)
			}
			ci.author = sig
		case "committer":
			sig, err := NewSignatureFromBytes(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse signature [%s]: %w", string(kv[1]), err)
			}
			ci.committer = sig
		case "gpgsig":
			begin := string(kv[1]) + "\n"
			end := "-----END PGP SIGNATURE-----"
			i := bytes.Index(objData[offset:], []byte(end))
			ci.gpgSig = begin + string(objData[offset:offset+i]) + end
			offset += len(end) + i + 1 // +1 to count the \n
		}
	}

	return ci, nil
}

// AsTag parses the object as Tag
//
// A tag has following format:
//
// object {sha}
// type {target_object_type}
// tag {tag_name}
// tagger {author_name} <{author_email}> {author_date_seconds} {author_date_timezone}
// gpgsig -----BEGIN PGP SIGNATURE-----
// {gpg key over multiple lines}
//  -----END PGP SIGNATURE-----
// {a blank line}
// {tag message}
//
// Note:
// - The gpgsig is optional
func (o *Object) AsTag() (*Tag, error) {
	if o.typ != TypeTag {
		return nil, xerrors.Errorf("type %s is not a tag", o.typ)
	}
	tag := &Tag{
		id:        o.ID(),
		rawObject: o,
	}
	offset := 0
	objData := o.Bytes()
	for {
		line := readutil.ReadTo(objData[offset:], '\n')
		offset += len(line) + 1 // +1 to count the \n

		// If we didn't find anything then something is wrong
		if len(line) == 0 && offset == 1 {
			return nil, xerrors.Errorf("could not find tag first line: %w", ErrTagInvalid)
		}

		// if we got an empty line, it means everything from now to the end
		// will be the tag message
		if len(line) == 0 {
			tag.message = string(objData[offset:])
			break
		}

		// Otherwise we're getting a key/value pair, separated by a space
		kv := bytes.SplitN(line, []byte{' '}, 2)
		switch string(kv[0]) {
		case "object":
			oid, err := ginternals.NewOidFromChars(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse target id %#v: %w", kv[1], err)
			}
			tag.target = oid
		case "type":
			typ, err := NewTypeFromString(string(kv[1]))
			if err != nil {
				return nil, xerrors.Errorf("object type %s: %w", string(kv[1]), err)
			}
			tag.typ = typ
		case "tagger":
			sig, err := NewSignatureFromBytes(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse signature [%s]: %w", string(kv[1]), err)
			}
			tag.tagger = sig
		case "tag":
			tag.tag = string(kv[1])
		case "gpgsig":
			begin := string(kv[1]) + "\n"
			end := "-----END PGP SIGNATURE-----"
			i := bytes.Index(objData[offset:], []byte(end))
			tag.gpgSig = begin + string(objData[offset:offset+i]) + end
			offset += len(end) + i + 1 // +1 to count the \n
		}
	}

	return tag, nil
}
