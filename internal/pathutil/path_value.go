package pathutil

import (
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
)

// PathValueType represents the type of a path
type PathValueType int

const (
	// PathValueTypeFile represent file
	PathValueTypeFile PathValueType = iota
	// PathValueTypeDir represent a directory
	PathValueTypeDir
	// PathValueTypeAny represent a either a file or a directory
	PathValueTypeAny
)

// PathValue represents a Flag value to be parsed by spf13/pflag
type PathValue struct {
	defaultValue  string
	userValue     string
	typ           PathValueType
	pathMustExist bool
	valueSet      bool
}

// NewDirPathFlagWithDefault return a new Flag Value that should hold
// a valid path to a directory
func NewDirPathFlagWithDefault(defaultPath string) pflag.Value {
	return &PathValue{
		pathMustExist: true,
		typ:           PathValueTypeDir,
		defaultValue:  defaultPath,
	}
}

// NewFilePathFlagWithDefault return a new Flag Value that should hold
// a valid path to a file
func NewFilePathFlagWithDefault(defaultPath string) pflag.Value {
	return &PathValue{
		pathMustExist: true,
		typ:           PathValueTypeFile,
		defaultValue:  defaultPath,
	}
}

// NewPathFlagWithDefault return a new Flag Value that should hold
// a valid path to either a file or a directory
func NewPathFlagWithDefault(defaultPath string) pflag.Value {
	return &PathValue{
		pathMustExist: true,
		typ:           PathValueTypeAny,
		defaultValue:  defaultPath,
	}
}

// we make sure the struct implements the interface
var _ pflag.Value = (*PathValue)(nil)

// String returns the flag's value
func (v *PathValue) String() string {
	if v.valueSet {
		return v.userValue
	}
	return v.defaultValue
}

// Set sets the flag's value.
// When called multiple times:
// - If the value is a relative path it will be append to the previous value
// - If the value is an absolute path: it will overwrite the previous value
func (v *PathValue) Set(value string) (err error) {
	if value == "" {
		return nil
	}

	if !filepath.IsAbs(value) {
		value = filepath.Join(v.userValue, value)
	}
	value, err = filepath.Abs(value)
	if err != nil {
		return xerrors.Errorf("could not find absolute path: %w", err)
	}

	info, err := os.Stat(value)
	if err != nil && !xerrors.Is(err, os.ErrNotExist) {
		return xerrors.Errorf("could not check path %s: %w", value, err)
	}

	if v.pathMustExist && xerrors.Is(err, os.ErrNotExist) {
		return xerrors.Errorf("invalid path %s: %w", value, os.ErrNotExist)
	}

	if info != nil {
		switch v.typ {
		case PathValueTypeFile:
			if info.IsDir() {
				return xerrors.Errorf("invalid path %s: is a directory", value)
			}
		case PathValueTypeDir:
			if !info.IsDir() {
				return xerrors.Errorf("invalid path %s: not a directory", value)
			}
		case PathValueTypeAny:
		default:
			return xerrors.Errorf("invalid type: %d", v.typ)
		}
	}

	v.valueSet = true
	v.userValue = value
	return nil
}

// Type returns the unique type of the Value
func (v *PathValue) Type() string {
	return "path"
}
