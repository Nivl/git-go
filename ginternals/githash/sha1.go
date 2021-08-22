package githash

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
)

// sha1NullOid represent an empty SHA1 OID
var sha1NullOid = sha1Oid{}

// sha1OidSize is the length of an oid, in bytes
const sha1OidSize = 20

// sha1Hash represents the sha1 implementation of the Hash interface
type sha1Hash struct{}

// NewSHA1 returns a Hash method using SHA1
func NewSHA1() Hash {
	return &sha1Hash{}
}

// OidSize returns the size of an Oid using this hash method
func (h *sha1Hash) OidSize() int {
	return sha1OidSize
}

// Sum returns the Oid of the given content.
// The oid will be the sum of the content
func (h *sha1Hash) Sum(bytes []byte) Oid {
	var oid sha1Oid = sha1.Sum(bytes)
	return oid
}

// ConvertFromString returns an Oid from the given string
// For the SHA 9b91da06e69613397b38e0808e0ba5ee6983251b
// the oid will be {0x9b, 0x91, 0xda, ...}
func (h *sha1Hash) ConvertFromString(id string) (Oid, error) {
	bytes, err := hex.DecodeString(id)
	if err != nil {
		if errors.Is(err, hex.ErrLength) {
			return sha1NullOid, ErrInvalidOid
		}
		return sha1NullOid, err
	}
	return h.ConvertFromBytes(bytes)
}

// ConvertFromString returns an Oid from the given char bytes
// For the SHA {'9', 'b', '9', '1', 'd', 'a', ...}
// the oid will be {0x9b, 0x91, 0xda, ...}
func (h *sha1Hash) ConvertFromChars(id []byte) (Oid, error) {
	return h.ConvertFromString(string(id))
}

// ConvertFromBytes returns an Oid from the provided byte-encoded oid
// This basically cast a slice that contains an encoded oid into
// a Oid object
func (h *sha1Hash) ConvertFromBytes(id []byte) (Oid, error) {
	if len(id) != sha1OidSize {
		return sha1NullOid, ErrInvalidOid
	}

	var oid sha1Oid
	copy(oid[:], id)
	return oid, nil
}

// NullOid returns an empty Oid
func (h *sha1Hash) NullOid() Oid {
	return sha1NullOid
}

// Name returns the name of the hash
func (h *sha1Hash) Name() string {
	return "sha1"
}

// sha1Oid represents an OID using SHA1
type sha1Oid [sha1OidSize]byte

// Bytes returns the raw Oid as []byte.
// This is different than doing []byte(oid.String())
// For the oid 642480605b8b0fd464ab5762e044269cf29a60a3:
// oid.Bytes(): []byte{ 0x64, 0x24, 0x80, ... }
// []byte(oid.String()): []byte{ '6', '4', '2', '4', '8' '0', ... }
func (o sha1Oid) Bytes() []byte {
	return o[:]
}

// String converts an oid to a string
func (o sha1Oid) String() string {
	return hex.EncodeToString(o[:])
}

// IsZero returns whether the oid has the zero value (NullOid)
func (o sha1Oid) IsZero() bool {
	return o == sha1NullOid
}
