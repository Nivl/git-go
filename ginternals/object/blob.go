package object

import "github.com/Nivl/git-go/ginternals/githash"

// Blob represents a blob object
type Blob struct {
	rawObject *Object
}

// NewBlob returns a new Blob object from a git Object
func NewBlob(o *Object) *Blob {
	return &Blob{
		rawObject: o,
	}
}

// ID returns the blob's ID
func (b *Blob) ID() githash.Oid {
	return b.rawObject.id
}

// Bytes returns the blob's contents
func (b *Blob) Bytes() []byte {
	return b.rawObject.content
}

// BytesCopy returns a copy of blob's contents
func (b *Blob) BytesCopy() []byte {
	out := make([]byte, len(b.rawObject.content))
	copy(out, b.rawObject.content)
	return out
}

// Size returns the size of the blob
func (b *Blob) Size() int {
	return len(b.rawObject.content)
}

// ToObject returns the Blob's underlying Object
func (b *Blob) ToObject() *Object {
	return b.rawObject
}
