package object

import (
	"bytes"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/readutil"
	"golang.org/x/xerrors"
)

// TagParams represents all the data needed to create a Tag
// Params starting by Opt are optionals
type TagParams struct {
	Target    *Object
	Name      string
	Tagger    Signature
	Message   string
	OptGPGSig string
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
func NewTag(p *TagParams) *Tag {
	return &Tag{
		target:  p.Target.ID(),
		typ:     p.Target.Type(),
		tag:     p.Name,
		tagger:  p.Tagger,
		message: p.Message,
		gpgSig:  p.OptGPGSig,
	}
}

// NewTagFromObject creates a new Tag from a raw git object
//
// A tag has following format:
//
// object {sha}
// type {target_object_type}
// tag {tag_name}
// tagger {author_name} <{author_email}> {author_date_seconds} {author_date_timezone}
// gpgsig -----BEGIN PGP SIGNATURE-----
// {gpg key over multiple lines}
//  -----END PGP SIGNATURE-----
// {a blank line}
// {tag message}
//
// Note:
// - The gpgsig is optional
func NewTagFromObject(o *Object) (*Tag, error) {
	if o.typ != TypeTag {
		return nil, xerrors.Errorf("type %s is not a tag", o.typ)
	}
	tag := &Tag{
		id:        o.ID(),
		rawObject: o,
	}
	offset := 0
	objData := o.Bytes()
	for {
		line := readutil.ReadTo(objData[offset:], '\n')
		offset += len(line) + 1 // +1 to count the \n

		// If we didn't find anything then something is wrong
		if len(line) == 0 && offset == 1 {
			return nil, xerrors.Errorf("could not find tag first line: %w", ErrTagInvalid)
		}

		// if we got an empty line, it means everything from now to the end
		// will be the tag message
		if len(line) == 0 {
			tag.message = string(objData[offset:])
			break
		}

		// Otherwise we're getting a key/value pair, separated by a space
		kv := bytes.SplitN(line, []byte{' '}, 2)
		switch string(kv[0]) {
		case "object":
			oid, err := ginternals.NewOidFromChars(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse target id %#v: %w", kv[1], err)
			}
			tag.target = oid
		case "type":
			typ, err := NewTypeFromString(string(kv[1]))
			if err != nil {
				return nil, xerrors.Errorf("object type %s: %w", string(kv[1]), err)
			}
			tag.typ = typ
		case "tagger":
			sig, err := NewSignatureFromBytes(kv[1])
			if err != nil {
				return nil, xerrors.Errorf("could not parse signature [%s]: %w", string(kv[1]), err)
			}
			tag.tagger = sig
		case "tag":
			tag.tag = string(kv[1])
		case "gpgsig":
			begin := string(kv[1]) + "\n"
			end := "-----END PGP SIGNATURE-----"
			i := bytes.Index(objData[offset:], []byte(end))
			tag.gpgSig = begin + string(objData[offset:offset+i]) + end
			offset += len(end) + i + 1 // +1 to count the \n
		}
	}

	return tag, nil
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
