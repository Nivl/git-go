package plumbing

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
)

const (
	// OidSize is the length of an oid, in bytes
	OidSize = 20
)

var (
	// NullOid is the value of an empty Oid, or one that's all 0s
	NullOid = Oid{}

	// ErrInvalidOid is returned when a given value isn't a valid Oid
	ErrInvalidOid = errors.New("invalid Oid")
)

// Oid represents an object id
type Oid [OidSize]byte

// Bytes returns the raw Oid as []byte.
// This is different than doing []byte(oid.String())
// For the oid 642480605b8b0fd464ab5762e044269cf29a60a3:
// oid.Bytes(): []byte{ 0x64, 0x24, 0x80, ... }
// []byte(oid.String()): []byte{ '6', '4', '2', '4', '8' '0', ... }
func (o Oid) Bytes() []byte {
	return o[:]
}

// String converts an oid to a string
func (o Oid) String() string {
	return hex.EncodeToString(o[:])
}

// NewOidFromContent returns the Oid of the given content.
// The oid will be the SHA1 sum of the content
func NewOidFromContent(bytes []byte) Oid {
	return sha1.Sum(bytes)
}

// NewOidFromHex returns an Oid from the provided byte-encoded oid
func NewOidFromHex(id []byte) (Oid, error) {
	if len(id) < OidSize {
		return NullOid, ErrInvalidOid
	}

	var oid Oid
	copy(oid[:], id)
	return oid, nil
}

// NewOidFromChars creates an Oid from the given char bytes
// For the SHA {'9', 'b', '9', '1', 'd', 'a', ...}
// the oid will be {0x9b, 0x91, 0xda, ...}
func NewOidFromChars(id []byte) (Oid, error) {
	return NewOidFromStr(string(id))
}

// NewOidFromStr creates an Oid from the given string
// For the SHA 9b91da06e69613397b38e0808e0ba5ee6983251b
// the oid will be {0x9b, 0x91, 0xda, ...}
func NewOidFromStr(id string) (Oid, error) {
	bytes, err := hex.DecodeString(id)
	if err != nil {
		return NullOid, err
	}

	if len(bytes) != 20 {
		return NullOid, ErrInvalidOid
	}

	var oid Oid
	copy(oid[:], bytes)

	return oid, nil
}

// IsZero returns whether the oid has the zero value (NullOid)
func (o Oid) IsZero() bool {
	return o == NullOid
}
