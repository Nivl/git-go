package fsbackend

import (
	"compress/zlib"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/readutil"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/Nivl/git-go/plumbing/packfile"
	"golang.org/x/xerrors"
)

// Object returns the object that has given oid
func (b *Backend) Object(oid plumbing.Oid) (*object.Object, error) {
	// First let's look for loose objects
	o, err := b.looseObject(oid)
	if err == nil {
		return o, nil
	}
	if !xerrors.Is(err, os.ErrNotExist) {
		return nil, xerrors.Errorf("failed looking for loose object: %w", err)
	}

	// Not found? Let's find it in a packfile
	o, err = b.objectFromPackfile(oid)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// looseObjectPath returns the absolute path of an object
// .git/object/first_2_chars_of_sha/remaining_chars_of_sha
// Ex. path of fcfe68a0e44e04bd7fd564fc0b75f1ae457e18b3 is:
// .git/objects/fc/fe68a0e44e04bd7fd564fc0b75f1ae457e18b3
func (b *Backend) looseObjectPath(sha string) string {
	return filepath.Join(b.root, gitpath.ObjectsPath, sha[:2], sha[2:])
}

// looseObject returns the object matching the given OID
// The format of an object is an ascii encoded type, an ascii encoded
// space, then an ascii encoded length of the object, then a null
// character, then the body of the object
func (b *Backend) looseObject(oid plumbing.Oid) (*object.Object, error) {
	strOid := oid.String()

	p := b.looseObjectPath(strOid)
	f, err := os.Open(p)
	if err != nil {
		return nil, xerrors.Errorf("could not find object %s at path %s: %w", strOid, p, err)
	}
	defer func() {
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// Objects are zlib encoded
	zlibReader, err := zlib.NewReader(f)
	if err != nil {
		return nil, xerrors.Errorf("could not decompress parts of object %s at path %s: %w", strOid, p, err)
	}
	defer func() {
		closeErr := zlibReader.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// We directly read the entire file since most of it is the content we
	// need, this allows us to be able to easily store the object's content
	buff, err := ioutil.ReadAll(zlibReader)
	if err != nil {
		return nil, xerrors.Errorf("could not read object %s at path %s: %w", strOid, p, err)
	}

	// we keep track of where we're at in the buffer
	pointerPos := 0

	// the type of the object starts at offset 0 and ends a the first
	// space character that we'll need to trim
	typ := readutil.ReadTo(buff, ' ')
	if typ == nil {
		return nil, xerrors.Errorf("could not find object type for %s at path %s: %w", strOid, p, err)
	}

	oType, err := object.NewTypeFromString(string(typ))
	if err != nil {
		return nil, xerrors.Errorf("unsupported type %s for object %s at path %s", string(typ), strOid, p)
	}
	pointerPos += len(typ)
	pointerPos++ // one more for the space

	// The size of the object starts after the space and ends at a NULL char
	// That we'll need to trim.
	// A NULL char is represented by 0 (dec), 000 (octal), or 0x00 (hex)
	// type "man ascii" in a terminal for more information
	size := readutil.ReadTo(buff[pointerPos:], 0)
	if size == nil {
		return nil, xerrors.Errorf("could not find object size for %s at path %s: %w", strOid, p, err)
	}
	oSize, err := strconv.Atoi(string(size))
	if err != nil {
		return nil, xerrors.Errorf("invalid size %s for object %s at path %s: %w", size, strOid, p, err)
	}
	pointerPos += len(size)
	pointerPos++                  // one more for the NULL char
	oContent := buff[pointerPos:] // sugar

	if len(oContent) != oSize {
		return nil, xerrors.Errorf("object marked as size %d, but has %d at path %s: %w", oSize, len(oContent), p, err)
	}

	return object.NewWithID(oid, oType, oContent), nil
}

// objectFromPackfile looks for an object in the packfiles
func (b *Backend) objectFromPackfile(oid plumbing.Oid) (*object.Object, error) {
	p := filepath.Join(b.root, gitpath.ObjectsPackPath)

	// TODO(melvin): parse MIDX files instead
	// MIDX file: https://git-scm.com/docs/multi-pack-index
	packfiles := []string{}
	err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// in case of error we just skip it and move on.
			return nil
		}

		if info.Name() == "pack" {
			return nil
		}

		// There should be no directories, but just in case,
		// we make sure we don't go in them
		if info.IsDir() {
			return filepath.SkipDir
		}

		// We're only interested in packfiles
		if filepath.Ext(info.Name()) != packfile.ExtPackfile {
			return nil
		}

		packfiles = append(packfiles, info.Name())
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, filename := range packfiles {
		// the index file of the packfile has the same name but with
		// the idx extension

		packFilePath := filepath.Join(p, filename)
		pf, err := packfile.NewFromFile(packFilePath)
		if err != nil {
			pf.Close() //nolint:errcheck // it failed anyway
			return nil, xerrors.Errorf("could not open packfile: %w", err)
		}
		do, err := pf.GetObject(oid)
		if err == nil {
			return do, pf.Close()
		}
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			if err = pf.Close(); err != nil {
				return nil, err
			}
			continue
		}
		pf.Close() //nolint:errcheck // it failed anyway
		return nil, err
	}
	return nil, plumbing.ErrObjectNotFound
}

// HasObject returns whether an object exists in the odb
func (b *Backend) HasObject(plumbing.Oid) (bool, error) {
	// TODO(melvin): add LRU cache so it's fine to call
	// HasObject() then Object() without perf hit
	panic("not implemented")
}

// WriteObject adds an object to the odb
func (b *Backend) WriteObject(o *object.Object) (plumbing.Oid, error) {
	data, err := o.Compress()
	if err != nil {
		return plumbing.NullOid, xerrors.Errorf("could not compress object: %w", err)
	}

	// Persist the data on disk
	sha := o.ID().String()
	p := b.looseObjectPath(sha)

	// TODO(melvin): Make sure the object doesn't already exist anywhere

	// We need to make sure the dest dir exists
	dest := filepath.Dir(p)
	if err = os.MkdirAll(dest, 0o755); err != nil {
		return plumbing.NullOid, xerrors.Errorf("could not create the destination directory %s: %w", dest, err)
	}

	if err = ioutil.WriteFile(p, data, 0o644); err != nil {
		return plumbing.NullOid, xerrors.Errorf("could not persist object %s at path %s: %w", sha, p, err)
	}

	return o.ID(), nil
}
