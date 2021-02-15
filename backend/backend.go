// Package backend contains interfaces and implementations to store and
// retrieve data from the odb
package backend

import (
	"errors"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/ginternals/packfile"
)

// This line generates a mock of the interfaces using gomock
// (https://github.com/golang/mock). To regenerate the mocks, you'll need
// gomock and mockgen installed, then run `go generate github.com/Nivl/git-go/backend`
//
//go:generate mockgen -package mockpackfile -destination ../internal/mocks/mockbackend/backend.go github.com/Nivl/git-go/backend Backend

// Backend represents an object that can store and retrieve data
// from and rto the odb
type Backend interface {
	// Close free the resources
	Close() error

	// Init initializes a repository
	Init() error

	// Reference returns a stored reference from its name
	Reference(name string) (*ginternals.Reference, error)
	// WriteReference writes the given reference int the db. If the
	// reference already exists it will be overwritten
	WriteReference(ref *ginternals.Reference) error
	// WriteReferenceSafe writes the given reference in the db
	// ErrRefExists is returned if the reference already exists
	WriteReferenceSafe(ref *ginternals.Reference) error
	// WalkReferences runs the provided method on all the references
	WalkReferences(f RefWalkFunc) error

	// Object returns the object that has given oid
	Object(ginternals.Oid) (*object.Object, error)
	// HasObject returns whether an object exists in the odb
	HasObject(ginternals.Oid) (bool, error)
	// WriteObject adds an object to the odb
	WriteObject(*object.Object) (ginternals.Oid, error)
	// WalkPackedObjectIDs runs the provided method on all the objects ids
	WalkPackedObjectIDs(f packfile.OidWalkFunc) error
	// WalkLooseObjectIDs runs the provided method on all the loose ids
	WalkLooseObjectIDs(f packfile.OidWalkFunc) error
}

// RefWalkFunc represents a function that will be applied on all references
// found by Walk()
type RefWalkFunc = func(ref *ginternals.Reference) error

// WalkStop is a fake error used to tell Walk() to stop
var WalkStop = errors.New("stop walking") //nolint // the linter expects all errors to start with Err, but since here we're faking an error we don't want that
