package object

import (
	"bytes"

	"github.com/Nivl/git-go/ginternals"
)

// TagOptions represents all the optional data available to create a Tag
type TagOptions struct {
	GPGSig string
}

// Tag represents a Tag object
type Tag struct {
	rawObject *Object

	tagger  Signature
	typ     Type
	tag     string
	message string

	gpgSig string

	id     ginternals.Oid
	target ginternals.Oid
}

// NewTag creates a new Tag object
func NewTag(target *Object, name string, tagger Signature, message string, opts TagOptions) (*Tag, error) {
	if target.ID().IsZero() {
		return nil, ErrObjectInvalid
	}
	c := &Tag{
		target:  target.ID(),
		typ:     target.Type(),
		tag:     name,
		tagger:  tagger,
		message: message,
		gpgSig:  opts.GPGSig,
	}
	return c, nil
}

// ID returns the SHA of the tag object
func (t *Tag) ID() ginternals.Oid {
	return t.id
}

// Target returns the ID of the object targeted by the tag
func (t *Tag) Target() ginternals.Oid {
	return t.target
}

// Type returns the type of the targeted object
func (t *Tag) Type() Type {
	return t.typ
}

// Name returns the tag's name
func (t *Tag) Name() string {
	return t.tag
}

// Tagger returns the Signature of the person that created the tag
func (t *Tag) Tagger() Signature {
	return t.tagger
}

// Message returns the tag's message
func (t *Tag) Message() string {
	return t.message
}

// GPGSig returns the GPG signature of the tag, if any
func (t *Tag) GPGSig() string {
	return t.gpgSig
}

// ToObject returns the underlying Object
func (t *Tag) ToObject() *Object {
	if t.rawObject != nil {
		return t.rawObject
	}

	// Quick reminder that the Write* methods on bytes.Buffer never fails,
	// the error returned is always nil
	buf := new(bytes.Buffer)
	buf.WriteString("object ")
	buf.WriteString(t.target.String())
	buf.WriteRune('\n')

	buf.WriteString("tag ")
	buf.WriteString(t.Name())
	buf.WriteRune('\n')

	buf.WriteString("type ")
	buf.WriteString(t.Type().String())
	buf.WriteRune('\n')

	buf.WriteString("tagger ")
	buf.WriteString(t.Tagger().String())
	buf.WriteRune('\n')

	if t.gpgSig != "" {
		buf.WriteString("gpgsig ")
		buf.WriteString(t.gpgSig)
		buf.WriteRune('\n')
	}

	buf.WriteRune('\n')

	buf.WriteString(t.message)
	t.rawObject = New(TypeTag, buf.Bytes())
	return t.rawObject
}
