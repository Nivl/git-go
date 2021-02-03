package testhelper

import (
	"github.com/spf13/pflag"
)

// StringValue represents a Flag value to be parsed by spf13/pflag
type StringValue struct {
	Value string
}

// NewStringValue creates a new StringValue Flag
func NewStringValue(v string) pflag.Value {
	return &StringValue{
		Value: v,
	}
}

// we make sure the struct implements the interface
var _ pflag.Value = (*StringValue)(nil)

// String returns the flag's value
func (v *StringValue) String() string {
	return v.Value
}

// Set sets the flag's value.
// When called multiple times:
// - If the value is a relative path it will be append to the previous value
// - If the value is an absolute path: it will overwrite the previous value
func (v *StringValue) Set(value string) (err error) {
	v.Value = value
	return nil
}

// Type returns the unique type of the Value
func (v *StringValue) Type() string {
	return "string"
}
