package git

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"strconv"

	"errors"

	"golang.org/x/xerrors"
)

// ErrObjectUnknown represents an error when encoutering an unknown object
var ErrObjectUnknown = errors.New("unknown object")

// ObjectType represents the type of an object as stored in a packfile
type ObjectType int8

// List of all the possible object types
const (
	ObjectTypeCommit ObjectType = 1
	ObjectTypeTree   ObjectType = 2
	ObjectTypeBlob   ObjectType = 3
	ObjectTypeTag    ObjectType = 4
	// 5 is reserved for future use
	ObjectDeltaOFS ObjectType = 6
	ObjectDeltaRef ObjectType = 7
)

func (t ObjectType) String() string {
	switch t {
	case ObjectTypeCommit:
		return "commit"
	case ObjectTypeTree:
		return "tree"
	case ObjectTypeBlob:
		return "blob"
	case ObjectTypeTag:
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
func (t ObjectType) IsValid() bool {
	switch t {
	case ObjectTypeCommit,
		ObjectTypeTree,
		ObjectTypeBlob,
		ObjectTypeTag,
		ObjectDeltaOFS,
		ObjectDeltaRef:
		return true
	default:
		return false
	}
}

// NewObjectTypeFromString returns an ObjectType from its string
// representation
func NewObjectTypeFromString(t string) (ObjectType, error) {
	switch t {
	case "commit":
		return ObjectTypeCommit, nil
	case "tree":
		return ObjectTypeTree, nil
	case "blob":
		return ObjectTypeBlob, nil
	case "tag":
		return ObjectTypeTag, nil
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
	ID      Oid
	typ     ObjectType
	size    int
	content []byte
}

// NewObject creates a new git object of the given type
// The Object ID won't be calculated until Compress() is called
func NewObject(typ ObjectType, content []byte) *Object {
	return &Object{
		ID:      NullOid,
		typ:     typ,
		size:    len(content),
		content: content,
	}
}

// Size returns the size of the object
func (o *Object) Size() int {
	return o.size
}

// Type returns the ObjectType for this object
func (o *Object) Type() ObjectType {
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
func (o *Object) Compress() (oid Oid, data []byte, err error) {
	// Quick reminder that the Write* methods on bytes.Buffer never fails,
	// the error returned is always nil
	w := new(bytes.Buffer)

	// Write the type
	w.WriteString(o.Type().String())
	// add the space
	w.WriteRune(' ')
	// write the size
	w.WriteString(strconv.Itoa(o.size))
	// Write the NULL char
	w.WriteByte(0)
	// Write the content
	w.Write(o.Bytes())

	// get the SHA of the file
	fileContent := w.Bytes()
	o.ID = NewOid(fileContent)

	compressedContent := new(bytes.Buffer)
	zw := zlib.NewWriter(compressedContent)
	defer func() {
		closeErr := zw.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if _, err = zw.Write(fileContent); err != nil {
		return NullOid, nil, xerrors.Errorf("could not zlib the object: %w", err)
	}
	if err = zw.Close(); err != nil {
		return NullOid, nil, xerrors.Errorf("could not close the compressor: %w", err)
	}
	return o.ID, compressedContent.Bytes(), nil
}

// AsBlob parses the object as Blob
func (o *Object) AsBlob() *Blob {
	return &Blob{Object: o}
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
	if o.typ != ObjectTypeCommit {
		return nil, xerrors.Errorf("type %s is not a commit", o.typ)
	}
	ci := &Commit{ID: o.ID}
	offset := 0
	objData := o.Bytes()
	for {
		line := readTo(objData[offset:], '\n')
		offset += len(line) + 1 // +1 to count the \n

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
			oid, err := NewOidFromBytes(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse tree id %#v: %w", kv[1], err)
			}
			ci.TreeID = oid
		case "parent":
			oid, err := NewOidFromBytes(kv[1])
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
			ci.gpgSig = begin + string(objData[offset:offset+i]) + end
			offset += len(end) + i
		}
	}

	return ci, nil
}
