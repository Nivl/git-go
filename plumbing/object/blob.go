package object

import "github.com/Nivl/git-go/plumbing"

// Blob represents a blob object
type Blob struct {
	ID   plumbing.Oid
	data []byte
}

// NewBlob returns a new blob from an object
func NewBlob(id plumbing.Oid, data []byte) *Blob {
	return &Blob{
		ID:   id,
		data: data,
	}
}

// Bytes returns a copy of blob's contents
func (b *Blob) Bytes() []byte {
	out := make([]byte, len(b.data))
	copy(out, b.data)
	return out
}

// Size returns the size of the blob
func (b *Blob) Size() int {
	return len(b.data)
}
