package fsbackend

import (
	"compress/zlib"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/ginternals/packfile"
	"github.com/Nivl/git-go/internal/errutil"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/readutil"
	"github.com/spf13/afero"
	"golang.org/x/xerrors"
)

// Object returns the object that has given oid
// This method can be called concurrently
func (b *Backend) Object(oid ginternals.Oid) (*object.Object, error) {
	key := oid[:]
	b.objectMu.Lock(key)
	defer b.objectMu.Unlock(key)

	return b.objectUnsafe(oid)
}

func (b *Backend) objectUnsafe(oid ginternals.Oid) (*object.Object, error) {
	if b.cache != nil {
		if cachedO, found := b.cache.Get(oid); found {
			if o, valid := cachedO.(*object.Object); valid {
				return o, nil
			}
		}
	}

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
	if b.cache != nil {
		b.cache.Add(oid, o)
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
// TODO(melvin): Move to ginternals (NewFromLoose or something)
func (b *Backend) looseObject(oid ginternals.Oid) (o *object.Object, err error) {
	if _, exists := b.looseObjects.Load(oid); !exists {
		return nil, os.ErrNotExist
	}

	strOid := oid.String()
	p := b.looseObjectPath(strOid)
	f, err := b.fs.Open(p)
	if err != nil {
		return nil, xerrors.Errorf("could not get object %s at path %s: %w", strOid, p, err)
	}
	defer errutil.Close(f, &err)

	// Objects are zlib encoded
	zlibReader, err := zlib.NewReader(f)
	if err != nil {
		return nil, xerrors.Errorf("could not decompress parts of object %s at path %s: %w", strOid, p, err)
	}
	defer errutil.Close(zlibReader, &err)

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

	return object.New(oType, oContent), nil
}

// loadPacks loads the packfiles in memory
func (b *Backend) loadPacks() error {
	p := filepath.Join(b.root, gitpath.ObjectsPackPath)
	return afero.Walk(b.fs, p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// in case of error we just skip it and move on.
			// this will happen if the repo is empty and the ./objects/pack
			// folder doesn't exists
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

		packFilePath := filepath.Join(p, info.Name())
		pack, err := packfile.NewFromFile(b.fs, packFilePath)
		if err != nil {
			return xerrors.Errorf("could not parse packfile at %s: %w", packFilePath, err)
		}
		b.packfiles[pack.ID()] = pack

		return nil
	})
}

// objectFromPackfile looks for an object in the packfiles
func (b *Backend) objectFromPackfile(oid ginternals.Oid) (*object.Object, error) {
	// TODO(melvin): parse MIDX files to speed up the process
	// MIDX file: https://git-scm.com/docs/multi-pack-index
	// https://github.com/Nivl/git-go/issues/13
	for _, pack := range b.packfiles {
		o, err := pack.GetObject(oid)
		if err == nil {
			return o, nil
		}
		if errors.Is(err, ginternals.ErrObjectNotFound) {
			continue
		}
		return nil, err
	}
	return nil, ginternals.ErrObjectNotFound
}

// HasObject returns whether an object exists in the odb
// This method can be called concurrently
func (b *Backend) HasObject(oid ginternals.Oid) (bool, error) {
	key := oid[:]
	b.objectMu.Lock(key)
	defer b.objectMu.Unlock(key)

	return b.hasObjectUnsafe(oid)
}

func (b *Backend) hasObjectUnsafe(oid ginternals.Oid) (bool, error) {
	_, err := b.objectUnsafe(oid)
	if err == nil {
		return true, nil
	}
	if xerrors.Is(err, ginternals.ErrObjectNotFound) {
		return false, nil
	}
	return false, xerrors.Errorf("could not get object: %w", err)
}

// WriteObject adds an object to the odb
// This method can be called concurrently
func (b *Backend) WriteObject(o *object.Object) (ginternals.Oid, error) {
	data, err := o.Compress()
	if err != nil {
		return ginternals.NullOid, xerrors.Errorf("could not compress object: %w", err)
	}

	oid := o.ID()
	b.objectMu.Lock(oid[:])
	defer b.objectMu.Unlock(oid[:])

	// Make sure the object doesn't already exist anywhere
	found, err := b.hasObjectUnsafe(o.ID())
	if err != nil {
		return ginternals.NullOid, xerrors.Errorf("could not check if object (%s) already exists: %w", o.ID().String(), err)
	}
	if found {
		return o.ID(), nil
	}

	// Persist the data on disk
	sha := o.ID().String()
	p := b.looseObjectPath(sha)

	// We need to make sure the dest dir exists
	dest := filepath.Dir(p)
	if err = b.fs.MkdirAll(dest, 0o755); err != nil {
		return ginternals.NullOid, xerrors.Errorf("could not create the destination directory %s: %w", dest, err)
	}

	// We use 444 because git object are read-only
	if err = afero.WriteFile(b.fs, p, data, 0o444); err != nil {
		return ginternals.NullOid, xerrors.Errorf("could not persist object %s at path %s: %w", sha, p, err)
	}

	// add the object to the cache
	b.looseObjects.Store(o.ID(), struct{}{})
	if b.cache != nil {
		b.cache.Add(o.ID(), o)
	}
	return o.ID(), nil
}

// WalkPackedObjectIDs runs the provided method on all the oids of all the
// packfiles
func (b *Backend) WalkPackedObjectIDs(f packfile.OidWalkFunc) error {
	for _, pack := range b.packfiles {
		if err := pack.WalkOids(f); err != nil {
			return err
		}
	}
	return nil
}

// loadLooseObject loads the loose object in memory
func (b *Backend) loadLooseObject() error {
	p := filepath.Join(b.root, gitpath.ObjectsPath)
	return afero.Walk(b.fs, p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// in case of error we just skip it and move on.
			// this will happen if the repo is empty and the ./objects
			// folder doesn't exists
			return nil
		}
		if path == p {
			return nil
		}

		// We're interested in all the directory that are named "00"
		// up to "ff"
		if info.IsDir() {
			if !b.isLooseObjectDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// We're only interested in the files inside a loose object
		// directory
		prefix := filepath.Base(filepath.Dir(path))
		if !b.isLooseObjectDir(prefix) {
			return filepath.SkipDir
		}

		if filepath.Ext(info.Name()) != "" {
			return filepath.SkipDir
		}

		sha := prefix + info.Name()
		oid, err := ginternals.NewOidFromStr(sha)
		if err != nil {
			return xerrors.Errorf("could not get oid from %s: %w", sha, err)
		}
		b.looseObjects.Store(oid, struct{}{})
		return nil
	})
}

// isLooseObjectDir checks if a directory name is anything between 00 and ff
func (b *Backend) isLooseObjectDir(name string) bool {
	if len(name) != 2 {
		return false
	}
	dirNum, parseErr := strconv.ParseInt(name, 16, 64)
	if parseErr != nil || dirNum < 0x00 || dirNum > 0xff {
		return false
	}
	return true
}

// WalkLooseObjectIDs runs the provided method on all the oids of all the
// packfiles
func (b *Backend) WalkLooseObjectIDs(f packfile.OidWalkFunc) (err error) {
	b.looseObjects.Range(func(key, value interface{}) bool {
		err = f(key.(ginternals.Oid))
		if err != nil {
			if err == packfile.OidWalkStop { //nolint:errorlint,goerr113 // it's a fake error so no need to use Error.Is()
				err = nil
			}
			return false
		}
		return true
	})
	return err
}
