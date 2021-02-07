package ginternals

import (
	"bytes"
	"errors"
	"strings"

	"golang.org/x/xerrors"
)

// Common ref names
const (
	// Head is a reference to the current branch, or to a commit if
	// we're detached
	Head = "HEAD"
	// OrigHead is a backup reference of HEAD set during destructive commands
	// such as rebase, merge, etc. and can be used to revert an operation
	OrigHead = "ORIG_HEAD"
	// MergeHead is a reference to the commit that is being merged
	// into the current branch
	MergeHead = "MERGE_HEAD"
	// CherryPickHead is a reference to the commit that is being
	// cherry-picked
	CherryPickHead = "CHERRY_PICK_HEAD"
	// Master correspond to the default branch name if none was
	// specified
	Master = "master"

	// FetchHead is a reference to the most recently fetched branch
	// TODO(melvin): Removed because the format is not currently
	// supported. It's a list of commit IDs with the branch name,
	// the origin, and other extra information. Example:
	//     bbb720a96e4c29b9950a4c577c98470a4d5dd089		branch 'master' of github.com:Nivl/git-go
	//     5f35f2dc6cec7356da02ca26192ce2bc3f271e79	not-for-merge	branch 'ml/feat/clone' of github.com:Nivl/git-go
	// FetchHead = "FETCH_HEAD"
)

var (
	// ErrRefNotFound is an error thrown when trying to act on a
	// reference that doesn't exists
	ErrRefNotFound = errors.New("reference not found")

	// ErrRefExists is an error thrown when trying to act on a
	// reference that should not exist, but does
	ErrRefExists = errors.New("reference already exists")

	// ErrRefNameInvalid is an error thrown when the name of a reference
	// is not valid
	ErrRefNameInvalid = errors.New("reference name is not valid")

	// ErrRefInvalid is an error thrown when a reference is not valid
	ErrRefInvalid = errors.New("reference is not valid")

	// ErrPackedRefInvalid is an error thrown when the packed-refs
	// file cannot be parsed properly
	ErrPackedRefInvalid = errors.New("packed-refs file is invalid")

	// ErrUnknownRefType is an error thrown when the type of a reference
	// is unknown
	ErrUnknownRefType = errors.New("unknown reference type")
)

// ReferenceType represents the type of a reference
type ReferenceType int8

const (
	// OidReference represents a reference that targets an Oid
	OidReference ReferenceType = 1
	// SymbolicReference represents a reference that targets another
	// reference
	SymbolicReference ReferenceType = 2
)

// Reference represents a git reference
// https://git-scm.com/book/en/v2/Git-Internals-Git-References
type Reference struct {
	name   string
	target string
	id     Oid
	typ    ReferenceType
}

// RefContent represents a method that returns the content of reference
// This is used so we can do the process here, without depending
// on a specific backend or having circular dependencies
type RefContent func(name string) ([]byte, error)

// ResolveReference resolves symbolic references
func ResolveReference(name string, finder RefContent) (*Reference, error) {
	return resolveRefs(name, finder, map[string]struct{}{})
}

// resolveRefs resolves references recursively
func resolveRefs(name string, finder RefContent, visited map[string]struct{}) (*Reference, error) {
	// we need to protect ourselves against circular references
	// Ex: refs/heads/master is a ref to refs/heads/a which is a ref to
	// refs/heads/master
	if _, ok := visited[name]; ok {
		return nil, xerrors.Errorf("circular symbolic reference: %w", ErrRefInvalid)
	}
	visited[name] = struct{}{}

	if !IsRefNameValid(name) {
		return nil, xerrors.Errorf(`ref "%s": %w`, name, ErrRefNameInvalid)
	}

	data, err := finder(name)
	if err != nil {
		return nil, err
	}
	data = bytes.Trim(data, " \n")

	// we're expecting at the very least 6 char:
	// "ref: " followed by a ref
	if len(data) < 6 {
		return nil, ErrRefInvalid
	}

	// if the reference is symbolic, we need to follow to get the target
	if string(data[0:5]) == "ref: " {
		symbolicTarget := string(data[5:])
		ref, err := resolveRefs(symbolicTarget, finder, visited)
		if err != nil {
			return nil, err
		}
		return &Reference{
			typ:    SymbolicReference,
			name:   name,
			id:     ref.id,
			target: symbolicTarget,
		}, nil
	}

	oid, err := NewOidFromChars(data)
	if err != nil {
		return nil, ErrRefInvalid
	}
	return &Reference{
		typ:  OidReference,
		name: name,
		id:   oid,
	}, nil
}

// NewReference return a new Reference object that targets
// an object
func NewReference(name string, target Oid) *Reference {
	return &Reference{
		typ:  OidReference,
		name: name,
		id:   target,
	}
}

// NewSymbolicReference return a new Reference object that targets
// another reference.
// Example HEAD targeting heads/master
func NewSymbolicReference(name, target string) *Reference {
	return &Reference{
		typ:    SymbolicReference,
		name:   name,
		target: target,
	}
}

// Name returns the full name fo the reference:
// example: refs/heads/master
func (ref *Reference) Name() string {
	return ref.name
}

// Target returns the ID targeted by a reference
func (ref *Reference) Target() Oid {
	return ref.id
}

// Type returns the type of a reference
func (ref *Reference) Type() ReferenceType {
	return ref.typ
}

// SymbolicTarget returns the symbolic target of a reference
func (ref *Reference) SymbolicTarget() string {
	return ref.target
}

// IsRefNameValid returns whether the name of a reference is valid or not
// https://stackoverflow.com/a/12093994/382879
func IsRefNameValid(name string) bool {
	// the reference name cannot:
	// - be empty
	// - start by a "/"
	// - end by a "/"
	// - end by .
	if name == "" || name == "/" || name[len(name)-1] == '/' || name[len(name)-1] == '.' {
		return false
	}

	// the reference name cannot contain:
	// - *
	// - ?
	// - ~
	// - :
	// - ^
	// - @{
	// - \
	// - ..
	// - [
	// - a space
	// - an ASCII char below 32 or a DEL (ASCII 127)
	for i, c := range name {
		if c < 32 || c == 127 {
			return false
		}
		if c == '*' || c == '?' || c == '!' || c == '^' {
			return false
		}
		if c == ' ' || c == '[' || c == '\\' || c == ':' {
			return false
		}
		if i < len(name)-1 {
			substr := name[i : i+2]
			if substr == "@{" || substr == ".." {
				return false
			}
		}
	}

	segments := strings.Split(name, "/")
	for _, s := range segments {
		// no segment cannot:
		// - be empty
		// - start by a dot
		// - end by a dot
		// - end by ".lock"
		if s == "" || s[0] == '.' || s[len(s)-1] == '.' || strings.HasSuffix(s, ".lock") {
			return false
		}
	}

	return true
}
