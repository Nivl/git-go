package ginternals

import "errors"

// ErrObjectNotFound is an error corresponding to a git object not being
// found
var ErrObjectNotFound = errors.New("object not found")
