package object

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/readutil"
)

// ErrSignatureInvalid is an error thrown when the signature of a commit
// couldn't be parsed
var ErrSignatureInvalid = errors.New("commit signature is invalid")

// Signature represents the author/committer and time of a commit
type Signature struct {
	Time  time.Time
	Name  string
	Email string
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
		if len(b) == 0 {
			return sig, fmt.Errorf("couldn't retrieve the name: %w", ErrSignatureInvalid)
		}
		return sig, fmt.Errorf("signature stopped after the name: %w", ErrSignatureInvalid)
	}
	sig.Name = strings.TrimSpace(string(data))
	offset := len(data) + 1 // +1 to skip the "<"
	if offset >= len(b) {
		if offset == len(b) {
			return sig, fmt.Errorf("couldn't retrieve the email: %w", ErrSignatureInvalid)
		}
		return sig, fmt.Errorf("signature stopped after the name: %w", ErrSignatureInvalid)
	}

	// Now we get the email, which is between "<" and ">"
	data = readutil.ReadTo(b[offset:], '>')
	if len(data) == 0 {
		return sig, fmt.Errorf("couldn't retrieve the email: %w", ErrSignatureInvalid)
	}
	sig.Email = string(data)
	// +2 to skip the "> "
	offset += len(data) + 2
	if offset >= len(b) {
		return sig, fmt.Errorf("signature stopped after the email: %w", ErrSignatureInvalid)
	}

	// Next is the timestamp and the timezone
	timestamp := readutil.ReadTo(b[offset:], ' ')
	if len(data) == 0 {
		// this should never be triggers since it's getting caught by the
		// previous check. Still leaving it to prevent introducing a bug
		// in the future.
		return sig, fmt.Errorf("couldn't retrieve the timestamp: %w", ErrSignatureInvalid)
	}
	offset += len(timestamp) + 1 // +1 to skip the " "
	if offset >= len(b) {
		return sig, fmt.Errorf("signature stopped after the timestamp: %w", ErrSignatureInvalid)
	}

	t, err := strconv.ParseInt(string(timestamp), 10, 64)
	if err != nil {
		return sig, fmt.Errorf("invalid timestamp %s: %w", timestamp, err)
	}
	sig.Time = time.Unix(t, 0)

	// To get and set the timezone we can just parse the time with an empty
	// date and copy it over to the signature
	timezone := b[offset:]
	tz, err := time.Parse("-0700", string(timezone))
	if err != nil {
		return sig, fmt.Errorf("invalid timezone format %s: %w", timezone, err)
	}
	sig.Time = sig.Time.In(tz.Location())
	return sig, nil
}

// CommitOptions represents all the optional data available to create a commit
type CommitOptions struct {
	Message string
	GPGSig  string
	// Committer represent the person creating the commit.
	// If not provided, the author will be used as committer
	Committer Signature
	ParentsID []ginternals.Oid
}

// Commit represents a commit object
type Commit struct {
	rawObject *Object

	author    Signature
	committer Signature

	gpgSig  string
	message string

	parentIDs []ginternals.Oid
	treeID    ginternals.Oid
}

// NewCommit creates a new Commit object
// Any provided Oids won't be check
func NewCommit(treeID ginternals.Oid, author Signature, opts *CommitOptions) *Commit {
	c := &Commit{
		treeID:    treeID,
		author:    author,
		committer: opts.Committer,
		message:   opts.Message,
		parentIDs: opts.ParentsID,
		gpgSig:    opts.GPGSig,
	}

	if c.committer.IsZero() {
		c.committer = author
	}
	c.rawObject = c.ToObject()

	return c
}

// NewCommitFromObject creates a commit from a raw object
//
// A commit has following format:
//
// tree {sha}
// parent {sha}
// author {author_name} <{author_email}> {author_date_seconds} {author_date_timezone}
// committer {committer_name} <{committer_email}> {committer_date_seconds} {committer_date_timezone}
// gpgsig -----BEGIN PGP SIGNATURE-----
// {gpg key over multiple lines}
//  -----END PGP SIGNATURE-----
// {a blank line}
// {commit message}
//
// Note:
// - A commit can have 0, 1, or many parents lines
//   The very first commit of a repo has no parents
//   A regular commit as 1 parent
//   A merge commit has 2 or more parents
// - The gpgsig is optional
func NewCommitFromObject(o *Object) (*Commit, error) {
	if o.typ != TypeCommit {
		return nil, fmt.Errorf("type %s is not a commit: %w", o.typ, ErrObjectInvalid)
	}
	ci := &Commit{
		rawObject: o,
	}
	offset := 0
	objData := o.Bytes()
	for {
		line := readutil.ReadTo(objData[offset:], '\n')
		offset += len(line) + 1 // +1 to count the \n

		// If we didn't find anything then something is wrong
		if len(line) == 0 && offset == 1 {
			return nil, fmt.Errorf("could not find commit first line: %w", ErrCommitInvalid)
		}

		// if we got an empty line, it means everything from now to the end
		// will be the commit message
		if len(line) == 0 {
			if offset < len(objData) {
				ci.message = string(objData[offset:])
			}
			break
		}

		// Otherwise we're getting a key/value pair, separated by a space
		kv := bytes.SplitN(line, []byte{' '}, 2)
		var err error
		switch string(kv[0]) {
		case "tree":
			ci.treeID, err = ginternals.NewOidFromChars(kv[1])
			if err != nil {
				return nil, fmt.Errorf("could not parse tree id %#v: %w", kv[1], err)
			}
		case "parent":
			oid, err := ginternals.NewOidFromChars(kv[1])
			if err != nil {
				return nil, fmt.Errorf("could not parse parent id %#v: %w", kv[1], err)
			}
			ci.parentIDs = append(ci.parentIDs, oid)
		case "author":
			ci.author, err = NewSignatureFromBytes(kv[1])
			if err != nil {
				return nil, fmt.Errorf("could not parse author signature [%s]: %w", string(kv[1]), err)
			}
		case "committer":
			ci.committer, err = NewSignatureFromBytes(kv[1])
			if err != nil {
				return nil, fmt.Errorf("could not parse committer signature [%s]: %w", string(kv[1]), err)
			}
		case "gpgsig":
			begin := string(kv[1]) + "\n"
			end := "-----END PGP SIGNATURE-----"
			i := bytes.Index(objData[offset:], []byte(end))
			ci.gpgSig = begin + string(objData[offset:offset+i]) + end
			offset += len(end) + i + 1 // +1 to count the \n
		}
	}

	// validate the commit
	if ci.author.IsZero() {
		return nil, fmt.Errorf("commit has no author: %w", ErrCommitInvalid)
	}
	if ci.treeID.IsZero() {
		return nil, fmt.Errorf("commit has no tree: %w", ErrCommitInvalid)
	}

	return ci, nil
}

// ID returns the SHA of the commit object
func (c *Commit) ID() ginternals.Oid {
	return c.rawObject.ID()
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
func (c *Commit) ParentIDs() []ginternals.Oid {
	out := make([]ginternals.Oid, len(c.parentIDs))
	copy(out, c.parentIDs)
	return out
}

// TreeID returns the SHA of the commit's tree
func (c *Commit) TreeID() ginternals.Oid {
	return c.treeID
}

// GPGSig returns the GPG signature of the commit, if any
func (c *Commit) GPGSig() string {
	return c.gpgSig
}

// ToObject returns the underlying Object
func (c *Commit) ToObject() *Object {
	if c.rawObject != nil {
		return c.rawObject
	}

	// Quick reminder that the Write* methods on bytes.Buffer never fails,
	// the error returned is always nil
	buf := new(bytes.Buffer)
	buf.WriteString("tree ")
	buf.WriteString(c.treeID.String())
	buf.WriteByte('\n')

	for _, p := range c.parentIDs {
		buf.WriteString("parent ")
		buf.WriteString(p.String())
		buf.WriteByte('\n')
	}

	buf.WriteString("author ")
	buf.WriteString(c.Author().String())
	buf.WriteByte('\n')

	buf.WriteString("committer ")
	buf.WriteString(c.Committer().String())
	buf.WriteByte('\n')

	if c.gpgSig != "" {
		buf.WriteString("gpgsig ")
		buf.WriteString(c.gpgSig)
		buf.WriteByte('\n')
	}

	buf.WriteByte('\n')

	buf.WriteString(c.message)
	return New(TypeCommit, buf.Bytes())
}
