package object

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Nivl/git-go/internal/readutil"
	"github.com/Nivl/git-go/plumbing"
	"golang.org/x/xerrors"
)

// ErrSignatureInvalid is an error thrown when the signature of a commit
// couldn't be parsed
var ErrSignatureInvalid = errors.New("commit signature is invalid")

// Signature represents the author/committer and time of a commit
type Signature struct {
	Name  string
	Email string
	Time  time.Time
}

// String returns a stringified version of the Signature
func (s Signature) String() string {
	return fmt.Sprintf("%s <%s> %d %s", s.Name, s.Email, s.Time.Unix(), s.Time.Format("-0700"))
}

// IsZero returns whether the signature has Zero value
func (s Signature) IsZero() bool {
	return s.Time.IsZero() && s.Name == "" && s.Email == ""
}

// NewSignature generates a signature at the current date and time
func NewSignature(name, email string) Signature {
	return Signature{
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
func NewSignatureFromBytes(b []byte) (Signature, error) {
	sig := Signature{}

	// First we get he name which will have the following format
	// "User Name " (with the extra space)
	data := readutil.ReadTo(b, '<')
	if len(data) == 0 {
		return sig, xerrors.Errorf("couldn't retrieve the name: %w", ErrSignatureInvalid)
	}
	sig.Name = strings.TrimSpace(string(data))
	offset := len(data) + 1 // +1 to skip the "<"
	if offset >= len(b) {
		return sig, xerrors.Errorf("signature stopped after the name: %w", ErrSignatureInvalid)
	}

	// Now we get the email, which is between "<" and ">"
	data = readutil.ReadTo(b[offset:], '>')
	if len(data) == 0 {
		return sig, xerrors.Errorf("couldn't retrieve the email: %w", ErrSignatureInvalid)
	}
	sig.Email = string(data)
	// +2 to skip the "> "
	offset += len(data) + 2
	if offset >= len(b) {
		return sig, xerrors.Errorf("signature stopped after the email: %w", ErrSignatureInvalid)
	}

	// Next is the timestamp and the timezone
	timestamp := readutil.ReadTo(b[offset:], ' ')
	if len(data) == 0 {
		return sig, xerrors.Errorf("couldn't retrieve the timestamp: %w", ErrSignatureInvalid)
	}
	offset += len(timestamp) + 1 // +1 to skip the " "
	if offset >= len(b) {
		return sig, xerrors.Errorf("signature stopped after the timestamp: %w", ErrSignatureInvalid)
	}

	t, err := strconv.ParseInt(string(timestamp), 10, 64)
	if err != nil {
		return sig, xerrors.Errorf("invalid timestamp %s: %w", timestamp, err)
	}
	sig.Time = time.Unix(t, 0)

	// To get and set the timezone we can just parse the time with an empty
	// date and copy it over to the signature
	timezone := b[offset:]
	tz, err := time.Parse("-0700", string(timezone))
	if err != nil {
		return sig, xerrors.Errorf("invalid timezone format %s: %w", timezone, err)
	}
	sig.Time = sig.Time.In(tz.Location())
	return sig, nil
}

// CommitOptions represents all the optional data available to create a commit
type CommitOptions struct {
	ParentsID []plumbing.Oid
	Message   string
	GPGSig    string
	// Committer represent the person creating the commit.
	// If not provided, the author will be used as committer
	Committer Signature
}

// Commit represents a commit object
type Commit struct {
	rawObject *Object

	author    Signature
	committer Signature

	gpgSig  string
	message string

	parentIDs []plumbing.Oid
	id        plumbing.Oid
	treeID    plumbing.Oid
}

// NewCommit creates a new Commit object
func NewCommit(tree *Tree, author Signature, opts *CommitOptions) *Commit {
	c := &Commit{
		treeID:    tree.ID(),
		author:    author,
		committer: opts.Committer,
		message:   opts.Message,
		parentIDs: opts.ParentsID,
		gpgSig:    opts.GPGSig,
	}

	if c.committer.IsZero() {
		c.committer = author
	}

	return c
}

// ID returns the SHA of the commit object
func (c *Commit) ID() plumbing.Oid {
	return c.id
}

// Author returns the Signature of the person that made the changes
func (c *Commit) Author() Signature {
	return c.author
}

// Committer returns the Signature of the person that created the commit
func (c *Commit) Committer() Signature {
	return c.committer
}

// Message returns the commit's message
func (c *Commit) Message() string {
	return c.message
}

// ParentIDs returns the list of SHA of the parent commits (if any)
// - The first commit of an orphan branch has 0 parents
// - A regular commit or the result of a fast-forward merge has 1 parent
// - A true merge (no fast-forward) as 2 or more parents
func (c *Commit) ParentIDs() []plumbing.Oid {
	out := make([]plumbing.Oid, len(c.parentIDs))
	copy(out, c.parentIDs)
	return out
}

// TreeID returns the SHA of the commit's tree
func (c *Commit) TreeID() plumbing.Oid {
	return c.treeID
}

// GPGSig returns the GPG signature of the commit, if any
func (c *Commit) GPGSig() string {
	return c.gpgSig
}

// ToObject returns the underlying Object
func (c *Commit) ToObject() (*Object, error) {
	if c.rawObject != nil {
		return c.rawObject, nil
	}

	// Quick reminder that the Write* methods on bytes.Buffer never fails,
	// the error returned is always nil
	buf := new(bytes.Buffer)
	buf.WriteString("tree ")
	buf.WriteString(c.treeID.String())
	buf.WriteRune('\n')

	for _, p := range c.parentIDs {
		buf.WriteString("parent ")
		buf.WriteString(p.String())
		buf.WriteRune('\n')
	}

	buf.WriteString("author ")
	buf.WriteString(c.Author().String())
	buf.WriteRune('\n')

	buf.WriteString("committer ")
	buf.WriteString(c.Committer().String())
	buf.WriteRune('\n')

	buf.WriteString("gpgsig ")
	buf.WriteString(c.gpgSig)
	buf.WriteRune('\n')

	buf.WriteRune('\n')

	buf.WriteString(c.message)
	if c.id != plumbing.NullOid {
		return NewWithID(c.id, TypeCommit, buf.Bytes()), nil
	}
	return New(TypeCommit, buf.Bytes()), nil
}
