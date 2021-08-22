package object

import (
	"bytes"
	"fmt"

	"github.com/Nivl/git-go/ginternals/githash"
	"github.com/Nivl/git-go/internal/readutil"
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
	tag     string
	message string

	gpgSig string

	id     githash.Oid
	target githash.Oid

	typ Type

	hash githash.Hash
}

// NewTag creates a new Tag object
func NewTag(hash githash.Hash, p *TagParams) *Tag {
	return &Tag{
		target:  p.Target.ID(),
		typ:     p.Target.Type(),
		tag:     p.Name,
		tagger:  p.Tagger,
		message: p.Message,
		gpgSig:  p.OptGPGSig,
		hash:    hash,
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
		return nil, fmt.Errorf("type %s is not a tag: %w", o.typ, ErrObjectInvalid)
	}
	tag := &Tag{
		id:        o.ID(),
		rawObject: o,
		hash:      o.hash,
	}
	offset := 0
	objData := o.Bytes()
	var err error
	for {
		line := readutil.ReadTo(objData[offset:], '\n')
		offset += len(line) + 1 // +1 to count the \n

		// If we didn't find anything then something is wrong
		if len(line) == 0 && offset == 1 {
			return nil, fmt.Errorf("could not find tag first line: %w", ErrTagInvalid)
		}

		// if we got an empty line, it means everything from now to the end
		// will be the tag message
		if len(line) == 0 {
			if offset < len(objData) {
				tag.message = string(objData[offset:])
			}
			break
		}

		// Otherwise we're getting a key/value pair, separated by a space
		kv := bytes.SplitN(line, []byte{' '}, 2)
		switch string(kv[0]) {
		case "object":
			tag.target, err = o.hash.ConvertFromChars(kv[1])
			if err != nil {
				return nil, fmt.Errorf("could not parse target id %#v: %w", kv[1], err)
			}
		case "type":
			tag.typ, err = NewTypeFromString(string(kv[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid object type %s: %w", string(kv[1]), err)
			}
		case "tagger":
			tag.tagger, err = NewSignatureFromBytes(kv[1])
			if err != nil {
				return nil, fmt.Errorf("could not parse tagger [%s]: %w", string(kv[1]), err)
			}
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

	// validate the tag
	if tag.tagger.IsZero() {
		return nil, fmt.Errorf("tag has no tagger: %w", ErrTagInvalid)
	}
	if tag.target.IsZero() {
		return nil, fmt.Errorf("tag has no target: %w", ErrTagInvalid)
	}
	if !tag.typ.IsValid() {
		return nil, fmt.Errorf("tag has no type: %w", ErrTagInvalid)
	}

	return tag, nil
}

// ID returns the SHA of the tag object
func (t *Tag) ID() githash.Oid {
	return t.id
}

// Target returns the ID of the object targeted by the tag
func (t *Tag) Target() githash.Oid {
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
	t.rawObject = New(t.hash, TypeTag, buf.Bytes())
	return t.rawObject
}
