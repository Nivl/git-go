// Package object contains methods and objects to work with git objects
package object

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/Nivl/git-go/ginternals/githash"
	"github.com/Nivl/git-go/internal/errutil"
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
	content      []byte
	idProcessing sync.Once
	typ          Type
	id           githash.Oid
	hash         githash.Hash
}

// New creates a new git object of the given type
func New(hash githash.Hash, typ Type, content []byte) *Object {
	o := &Object{
		typ:     typ,
		content: content,
		hash:    hash,
	}
	o.id, _ = o.build()
	return o
}

// ID returns the ID of the object.
func (o *Object) ID() githash.Oid {
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

func (o *Object) build() (oid githash.Oid, data []byte) {
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
	oid = o.hash.Sum(data)
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
		return nil, fmt.Errorf("could not zlib the object: %w", err)
	}
	return compressedContent.Bytes(), nil
}

// AsBlob parses the object as Blob
func (o *Object) AsBlob() *Blob {
	return NewBlob(o)
}

// AsTree parses the object as Tree
func (o *Object) AsTree() (*Tree, error) {
	return NewTreeFromObject(o)
}

// AsCommit parses the object as Commit
func (o *Object) AsCommit() (*Commit, error) {
	return NewCommitFromObject(o)
}

// AsTag parses the object as Tag
func (o *Object) AsTag() (*Tag, error) {
	return NewTagFromObject(o)
}
