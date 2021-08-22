package githash

import "errors"

// ErrInvalidOid is returned when a given value isn't a valid Oid
var ErrInvalidOid = errors.New("invalid Oid")

// Hash represents an Hash algorithm supported by Git
type Hash interface {
	// Name returns the name of the hash
	Name() string

	OidSize() int
	// Sum returns the Oid of the given content.
	// The oid will be the sum of the content
	Sum(bytes []byte) Oid
	// ConvertFromString returns an Oid from the given string
	// For the SHA 9b91da06e69613397b38e0808e0ba5ee6983251b
	// the oid will be {0x9b, 0x91, 0xda, ...}
	ConvertFromString(id string) (Oid, error)
	// ConvertFromString returns an Oid from the given char bytes
	// For the SHA {'9', 'b', '9', '1', 'd', 'a', ...}
	// the oid will be {0x9b, 0x91, 0xda, ...}
	ConvertFromChars(id []byte) (Oid, error)
	// ConvertFromBytes returns an Oid from the provided byte-encoded oid
	// This basically cast a slice that contains an encoded oid into
	// a Oid object
	ConvertFromBytes(id []byte) (Oid, error)
	// NullOid returns an empty Oid
	NullOid() Oid
}

// Oid represents a git Object ID
type Oid interface {
	// Bytes returns the raw Oid as []byte.
	// This is different than doing []byte(oid.String())
	// For the oid 642480605b8b0fd464ab5762e044269cf29a60a3:
	// oid.Bytes(): []byte{ 0x64, 0x24, 0x80, ... }
	// []byte(oid.String()): []byte{ '6', '4', '2', '4', '8' '0', ... }
	Bytes() []byte

	// String converts an oid to a string
	String() string

	// IsZero returns whether the oid has the zero value (NullOid)
	IsZero() bool
}
