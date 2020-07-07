package fsbackend

import (
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
)

// Object returns the object that has given oid
func (b *Backend) Object(plumbing.Oid) (plumbing.Oid, error) {
	return plumbing.NullOid, nil
}

// HasObject returns whether an object exists in the odb
func (b *Backend) HasObject(plumbing.Oid) (bool, error) {
	return false, nil
}

// WriteObject adds an object to the odb
func (b *Backend) WriteObject(*object.Object) (plumbing.Oid, error) {
	return plumbing.NullOid, nil
}
