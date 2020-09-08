package object

import (
	"bytes"

	"github.com/Nivl/git-go/ginternals"
)

// TagOptions represents all the optional data available to create a Tag
type TagOptions struct {
	GPGSig  string
	Message string
}

// Tag represents a Tag object
type Tag struct {
	rawObject *Object

	tagger  Signature
	typ     Type
	tag     string
	message string

	gpgSig string

	id       ginternals.Oid
	targetID ginternals.Oid
}

// NewTag creates a new Tag object
func NewTag(targetID ginternals.Oid, name string, tagger Signature, opts TagOptions) *Tag {
	c := &Tag{
		targetID: targetID,
		tag:      name,
		tagger:   tagger,
		message:  opts.Message,
		gpgSig:   opts.GPGSig,
	}
	return c
}

// ID returns the SHA of the tag object
func (c *Tag) ID() ginternals.Oid {
	return c.id
}

// TargetID returns the ID of the object targeted by the tag
func (c *Tag) TargetID() ginternals.Oid {
	return c.targetID
}

// Type returns the type of the targeted object
func (c *Tag) Type() Type {
	return c.typ
}

// Name returns the tag's name
func (c *Tag) Name() string {
	return c.tag
}

// Tagger returns the Signature of the person that created the tag
func (c *Tag) Tagger() Signature {
	return c.tagger
}

// Message returns the tag's message
func (c *Tag) Message() string {
	return c.message
}

// GPGSig returns the GPG signature of the tag, if any
func (c *Tag) GPGSig() string {
	return c.gpgSig
}

// ToObject returns the underlying Object
func (c *Tag) ToObject() *Object {
	if c.rawObject != nil {
		return c.rawObject
	}

	// Quick reminder that the Write* methods on bytes.Buffer never fails,
	// the error returned is always nil
	buf := new(bytes.Buffer)
	buf.WriteString("object ")
	buf.WriteString(c.targetID.String())
	buf.WriteRune('\n')

	buf.WriteString("tag ")
	buf.WriteString(c.Name())
	buf.WriteRune('\n')

	buf.WriteString("type ")
	buf.WriteString(c.Type().String())
	buf.WriteRune('\n')

	buf.WriteString("tagger ")
	buf.WriteString(c.Tagger().String())
	buf.WriteRune('\n')

	if c.gpgSig != "" {
		buf.WriteString("gpgsig ")
		buf.WriteString(c.gpgSig)
		buf.WriteRune('\n')
	}

	buf.WriteRune('\n')

	buf.WriteString(c.message)
	return New(TypeTag, buf.Bytes())
}
