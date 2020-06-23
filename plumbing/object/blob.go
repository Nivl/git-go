package object

import "github.com/Nivl/git-go/plumbing"

// Blob represents a blob object
type Blob struct {
	ID        plumbing.Oid
	rawObject *Object
}

// NewBlob returns a new blob from an object
func NewBlob(object *Object) *Blob {
	return &Blob{
		ID:        object.ID,
		rawObject: object,
	}
}

// Bytes returns the blob's contents
func (b *Blob) Bytes() []byte {
	return b.rawObject.Bytes()
}

// Size returns the size of the blob
func (b *Blob) Size() int {
	return b.rawObject.Size()
}
