package git

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Signature represents the author/committer and time of a commit
type Signature struct {
	Name  string
	Email string
	Time  time.Time
}

// NewSignature generates a signature at the current date and time
func NewSignature(name, email string) *Signature {
	return &Signature{
		Name:  name,
		Email: email,
		Time:  time.Now(),
	}
}

// NewSignatureFromBytes returns a signature from an array of byte
//
// A signature has the following format:
// User Name <user.email@domain.tld> timestamp timezone
// Ex:
// Melvin Laplanche <melvin.wont.reply@gmail.com> 1566115917 -0700
func NewSignatureFromBytes(b []byte) (*Signature, error) {
	sig := &Signature{}

	// First we get he name which will have the following format
	// "User Name " (with the extra space)
	data := readTo(b, '<')
	if len(data) == 0 {
		return nil, errors.New("couldn't retreive the name")
	}
	sig.Name = strings.TrimSpace(string(data))
	offset := len(data) + 1 // +1 to skip the "<"
	if offset >= len(b) {
		return nil, errors.New("signature stopped after the name")
	}

	// Now we get the email, which is between "<" and ">"
	data = readTo(b[offset:], '>')
	if len(data) == 0 {
		return nil, errors.New("couldn't retreive the email")
	}
	sig.Email = string(data)
	// +2 to skip the "> "
	offset += len(data) + 2
	if offset >= len(b) {
		return nil, errors.New("signature stopped after the email")
	}

	// Next is the timestamp and the timezone
	timestamp := readTo(b[offset:], ' ')
	if len(data) == 0 {
		return nil, errors.New("couldn't retreive the timestamp")
	}
	offset += len(timestamp) + 1 // +1 to skip the " "
	if offset >= len(b) {
		return nil, errors.New("signature stopped after the timestamp")
	}

	t, err := strconv.ParseInt(string(timestamp), 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid timestamp %s", timestamp)
	}
	sig.Time = time.Unix(t, 0)

	// To get and set the timezone we can just parse the time with an empty
	// date and copy it over to the signature
	timezone := b[offset:]
	tz, err := time.Parse("-0700", string(timezone))
	if err != nil {
		return nil, errors.Wrapf(err, "invalid timezone format %s", timezone)
	}
	sig.Time = sig.Time.In(tz.Location())
	return sig, nil
}

// Commit represents a commit object
type Commit struct {
	ID Oid

	// SHA of all the parent commits, if any
	// A regular commit usually has 1 parent, a merge commit has 2 or more,
	// and the very first commit has none
	ParentIDs []Oid

	TreeID    Oid
	Author    *Signature
	Committer *Signature
	gpgSig    string
	Message   string
}
