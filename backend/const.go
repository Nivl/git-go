package backend

import (
	"errors"

	"github.com/Nivl/git-go/ginternals"
)

// .git/Config config keys
const (
	CfgCore                  = "core"
	CfgCoreFormatVersion     = "repositoryformatversion"
	CfgCoreFileMode          = "filemode"
	CfgCoreBare              = "bare"
	CfgCoreLogAllRefUpdate   = "logallrefupdates"
	CfgCoreIgnoreCase        = "ignorecase"
	CfgCorePrecomposeUnicode = "precomposeunicode"
)

// RefWalkFunc represents a function that will be applied on all references
// found by Walk()
type RefWalkFunc = func(ref *ginternals.Reference) error

// WalkStop is a fake error used to tell Walk() to stop
var WalkStop = errors.New("stop walking") //nolint // the linter expects all errors to start with Err, but since here we're faking an error we don't want that
