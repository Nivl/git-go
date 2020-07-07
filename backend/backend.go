// Package backend contains interfaces and implementations to store and
// retrieve data from the odb
package backend

import (
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
)

// This line generates a mock of the interfaces using gomock
// (https://github.com/golang/mock). To regenerate the mocks, you'll need
// gomock and mockgen installed, then run `go generate github.com/Nivl/git-go/backend`
//
//go:generate mockgen -package mockpackfile -destination ../internal/mocks/mockbackend/backend.go github.com/Nivl/git-go/backend Backend

// Backend represents an object that can store and retrieve data
// from and rto the odb
type Backend interface {
	// Init initializes a repository
	Init() error

	// Reference returns a stored reference from its name
	Reference(name string) (*plumbing.Reference, error)
	// WriteReference writes the given reference int the db. If the
	// reference already exists it will be overwritten
	WriteReference(ref *plumbing.Reference) error
	// WriteReferenceSafe writes the given reference in the db
	// ErrRefExists is returned if the reference already exists
	WriteReferenceSafe(ref *plumbing.Reference) error

	// Object returns the object that has given oid
	Object(plumbing.Oid) (*object.Object, error)
	// HasObject returns whether an object exists in the odb
	HasObject(plumbing.Oid) (bool, error)
	// WriteObject adds an object to the odb
	WriteObject(*object.Object) (plumbing.Oid, error)
}