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

// Bytes returns a byte slice of the Oid
func (o Oid) Bytes() []byte {
	return o[:]
}

// String converts an oid to a string
func (o Oid) String() string {
	return hex.EncodeToString(o[:])
}

// NewOid returns the Oid of the given content. The oid will be the
// SHA1 sum of the content
func NewOid(bytes []byte) Oid {
	return sha1.Sum(bytes)
}

// NewOidFromBytes returns an Oid from the provided byte encoded oid
func NewOidFromBytes(id []byte) (Oid, error) {
	if len(id) < OidSize {
		return NullOid, ErrInvalidOid
	}

	var oid Oid
	copy(oid[:], id)
	return oid, nil
}

// NewOidFromStr creates an Oid from the given string
// For the SHA 9b91da06e69613397b38e0808e0ba5ee6983251b
// the oid will be {'9b', '91', 'da', ...}
func NewOidFromStr(id string) (Oid, error) {
	bytes, decodeErr := hex.DecodeString(id)
	if decodeErr != nil {
		return NullOid, decodeErr
	}

	if len(bytes) != 20 {
		return NullOid, ErrInvalidOid
	}

	var oid Oid
	copy(oid[:], bytes)

	return oid, nil
}
