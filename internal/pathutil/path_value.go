package pathutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
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

var (
	// ErrIsDirectory is an error returned when a path
	// points to a directory instead of a file
	ErrIsDirectory = errors.New("path is a directory")
	// ErrIsNotDirectory is an error returned when a path
	// is expected to points to a directory but isn't
	ErrIsNotDirectory = errors.New("path is not a directory")
	// ErrUnknownType is an error returned when an unknown PathValueType
	// is provided to a method
	ErrUnknownType = errors.New("type unknown")
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
		return fmt.Errorf("could not find absolute path: %w", err)
	}

	info, err := os.Stat(value)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not check path %s: %w", value, err)
	}

	if v.pathMustExist && errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("invalid path %s: %w", value, os.ErrNotExist)
	}

	if info != nil {
		switch v.typ {
		case PathValueTypeFile:
			if info.IsDir() {
				return fmt.Errorf("invalid path %s: %w", value, ErrIsDirectory)
			}
		case PathValueTypeDir:
			if !info.IsDir() {
				return fmt.Errorf("invalid path %s: %w", value, ErrIsNotDirectory)
			}
		case PathValueTypeAny:
		default:
			return fmt.Errorf("type %d: %w", v.typ, ErrUnknownType)
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
