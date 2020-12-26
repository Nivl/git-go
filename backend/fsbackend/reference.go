package fsbackend

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/gitpath"
	"golang.org/x/xerrors"
)

// Reference returns a stored reference from its name
// ErrRefNotFound is returned if the reference doesn't exists
func (b *Backend) Reference(name string) (*ginternals.Reference, error) {
	var packedRef map[string]string

	finder := func(name string) ([]byte, error) {
		data, err := ioutil.ReadFile(b.systemPath(name))
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, xerrors.Errorf("could not read reference content: %w", err)
			}
			// if the reference can't be found on disk, it might be
			// in the packed-ref file
			if packedRef == nil {
				packedRef, err = b.parsePackedRefs()
				if err != nil {
					return nil, xerrors.Errorf("couldn't load packed-refs: %w", err)
				}
			}
			sha, ok := packedRef[name]
			if !ok {
				return nil, xerrors.Errorf(`ref "%s": %w`, name, ginternals.ErrRefNotFound)
			}
			return []byte(sha), nil
		}
		return data, nil
	}
	return ginternals.ResolveReference(name, finder)
}

// systemPath returns a path from a ref name
// Ex.: On windows refs/heads/master would return refs\heads\master
func (b *Backend) systemPath(name string) string {
	switch os.PathSeparator {
	case '/':
		return filepath.Join(b.root, name)
	default:
		name = filepath.FromSlash(name)
		return filepath.Join(b.root, name)
	}
}

// parsePackedRefs parsed the packed-refs file and returns a map
// refName => Oid
// https://git-scm.com/docs/git-pack-refs
func (b *Backend) parsePackedRefs() (refs map[string]string, err error) {
	refs = map[string]string{}
	f, err := b.fs.Open(filepath.Join(b.root, gitpath.PackedRefsPath))
	if err != nil {
		// if the file doesn't exist we just return an empty map
		if os.IsNotExist(err) {
			return refs, nil
		}
		return nil, xerrors.Errorf("could not open %s: %w", gitpath.PackedRefsPath, err)
	}
	defer func() {
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}()

	sc := bufio.NewScanner(f)
	for i := 1; sc.Scan(); i++ {
		i++
		line := sc.Text()
		// we skip empty lines, comments, and annotated tag commit
		if line == "" || line[0] == '#' || line[0] == '^' {
			continue
		}
		// We expected data to have the format:
		// "oid ref-name"
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			return nil, xerrors.Errorf("unexpected data line %d: %w", i, ginternals.ErrPackedRefInvalid)
		}
		refs[parts[1]] = parts[0]
	}

	if sc.Err() != nil {
		return nil, xerrors.Errorf("could not parse %s: %w", gitpath.PackedRefsPath, err)
	}

	return refs, nil
}

// WriteReference writes the given reference on disk. If the
// reference already exists it will be overwritten
func (b *Backend) WriteReference(ref *ginternals.Reference) error {
	if !ginternals.IsRefNameValid(ref.Name()) {
		return ginternals.ErrRefNameInvalid
	}

	target := ""
	switch ref.Type() {
	case ginternals.SymbolicReference:
		target = fmt.Sprintf("ref: %s\n", ref.SymbolicTarget())
	case ginternals.OidReference:
		target = fmt.Sprintf("%s\n", ref.Target().String())
	default:
		return xerrors.Errorf("reference type %d: %w", ref.Type(), ginternals.ErrUnknownRefType)
	}
	err := ioutil.WriteFile(b.systemPath(ref.Name()), []byte(target), 0o644)
	if err != nil {
		return xerrors.Errorf("could not persist reference to disk: %w", err)
	}
	return nil
}

// WriteReferenceSafe writes the given reference in the db
// ErrRefExists is returned if the reference already exists
func (b *Backend) WriteReferenceSafe(ref *ginternals.Reference) error {
	if !ginternals.IsRefNameValid(ref.Name()) {
		return ginternals.ErrRefNameInvalid
	}

	// First we check if the reference is on disk
	p := b.systemPath(ref.Name())
	_, err := b.fs.Stat(p)
	if !os.IsNotExist(err) {
		if err != nil {
			return xerrors.Errorf("could not check if reference exists on disk: %w", err)
		}
		return ginternals.ErrRefExists
	}

	// Now we check if the reference is on the packed-refs file
	refs, err := b.parsePackedRefs()
	if err != nil {
		return xerrors.Errorf("could not check %s: %w", gitpath.PackedRefsPath, err)
	}
	if _, ok := refs[ref.Name()]; ok {
		return ginternals.ErrRefExists
	}

	return b.WriteReference(ref)
}
