package githash

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// sha256NullOid represent an empty SHA256 OID
var sha256NullOid = sha256Oid{}

// sha256OidSize is the length of an oid, in bytes
const sha256OidSize = sha256.Size

// sha256Hash represents the sha256 implementation of the Hash interface
type sha256Hash struct{}

// NewSHA256 returns a Hash method using SHA256
func NewSHA256() Hash {
	return &sha256Hash{}
}

// OidSize returns the size of an Oid using this hash method
func (h *sha256Hash) OidSize() int {
	return sha256OidSize
}

// Sum returns the Oid of the given content.
// The oid will be the sum of the content
func (h *sha256Hash) Sum(bytes []byte) Oid {
	var oid sha256Oid = sha256.Sum256(bytes)
	return oid
}

// ConvertFromString returns an Oid from the given string
// For the SHA 9b91da06e69613397b38e0808e0ba5ee6983251b
// the oid will be {0x9b, 0x91, 0xda, ...}
func (h *sha256Hash) ConvertFromString(id string) (Oid, error) {
	bytes, err := hex.DecodeString(id)
	if err != nil {
		if errors.Is(err, hex.ErrLength) {
			return sha256NullOid, ErrInvalidOid
		}
		return sha256NullOid, err
	}
	return h.ConvertFromBytes(bytes)
}

// ConvertFromString returns an Oid from the given char bytes
// For the SHA {'9', 'b', '9', '1', 'd', 'a', ...}
// the oid will be {0x9b, 0x91, 0xda, ...}
func (h *sha256Hash) ConvertFromChars(id []byte) (Oid, error) {
	return h.ConvertFromString(string(id))
}

// ConvertFromBytes returns an Oid from the provided byte-encoded oid
// This basically cast a slice that contains an encoded oid into
// a Oid object
func (h *sha256Hash) ConvertFromBytes(id []byte) (Oid, error) {
	if len(id) != sha256OidSize {
		return sha256NullOid, ErrInvalidOid
	}

	var oid sha256Oid
	copy(oid[:], id)
	return oid, nil
}

// NullOid returns an empty Oid
func (h *sha256Hash) NullOid() Oid {
	return sha256NullOid
}

// Name returns the name of the hash
func (h *sha256Hash) Name() string {
	return "sha256"
}

// sha256Oid represents an OID using SHA256
type sha256Oid [sha256OidSize]byte

// Bytes returns the raw Oid as []byte.
// This is different than doing []byte(oid.String())
// For the oid 642480605b8b0fd464ab5762e044269cf29a60a3:
// oid.Bytes(): []byte{ 0x64, 0x24, 0x80, ... }
// []byte(oid.String()): []byte{ '6', '4', '2', '4', '8' '0', ... }
func (o sha256Oid) Bytes() []byte {
	return o[:]
}

// String converts an oid to a string
func (o sha256Oid) String() string {
	return hex.EncodeToString(o[:])
}

// IsZero returns whether the oid has the zero value (NullOid)
func (o sha256Oid) IsZero() bool {
	return o == sha256NullOid
}
