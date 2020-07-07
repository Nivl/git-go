package fsbackend

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/plumbing"
	"golang.org/x/xerrors"
)

// Reference returns a stored reference from its name
// ErrRefNotFound is returned if the reference doesn't exists
func (b *Backend) Reference(name string) (*plumbing.Reference, error) {
	var packedRef map[string]string

	finder := func(name string) ([]byte, error) {
		data, err := ioutil.ReadFile(b.nameToPath(name))
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
				return nil, xerrors.Errorf(`ref "%s": %w`, name, plumbing.ErrRefNotFound)
			}
			return []byte(sha), nil
		}
		return data, nil
	}
	return plumbing.ResolveReference(name, finder)
}

// nameToPath returns a path from a ref name
// Ex.: On windows refs/heads/master would return refs\heads\master
func (b *Backend) nameToPath(name string) string {
	switch os.PathSeparator {
	case '/':
		return filepath.Join(b.root, name)
	default:
		parts := strings.Split(name, "/")
		p := filepath.Join(parts...)
		return filepath.Join(b.root, p)
	}
}

// parsePackedRefs parsed the packed-refs file and returns a map
// refName => Oid
// https://git-scm.com/docs/git-pack-refs
func (b *Backend) parsePackedRefs() (map[string]string, error) {
	refs := map[string]string{}
	f, err := os.Open(filepath.Join(b.root, gitpath.PackedRefsPath))
	if err != nil {
		// if the file doesn't exist we just return an empty map
		if os.IsNotExist(err) {
			return refs, nil
		}
		return nil, xerrors.Errorf("could not open %s: %w", gitpath.PackedRefsPath, err)
	}

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
			return nil, xerrors.Errorf("unexpected data line %d: %w", i, plumbing.ErrPackedRefInvalid)
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
func (b *Backend) WriteReference(ref *plumbing.Reference) error {
	if !plumbing.IsRefNameValid(ref.Name()) {
		return plumbing.ErrRefNameInvalid
	}

	target := ""
	switch ref.Type() {
	case plumbing.SymbolicReference:
		target = fmt.Sprintf("ref: %s\n", ref.SymbolicTarget())
	case plumbing.OidReference:
		target = fmt.Sprintf("%s\n", ref.Target().String())
	default:
		return xerrors.Errorf("reference type %d: %w", ref.Type(), plumbing.ErrUnknownRefType)
	}
	err := ioutil.WriteFile(b.nameToPath(ref.Name()), []byte(target), 0644)
	if err != nil {
		return xerrors.Errorf("could not persist reference to disk: %w", err)
	}
	return nil
}

// WriteReferenceSafe writes the given reference in the db
// ErrRefExists is returned if the reference already exists
func (b *Backend) WriteReferenceSafe(ref *plumbing.Reference) error {
	if !plumbing.IsRefNameValid(ref.Name()) {
		return plumbing.ErrRefNameInvalid
	}

	// First we check if the reference is on disk
	p := b.nameToPath(ref.Name())
	_, err := os.Stat(p)
	if !os.IsNotExist(err) {
		if err != nil {
			return xerrors.Errorf("could not check if reference exists on disk: %w", err)
		}
		return plumbing.ErrRefExists
	}

	// Now we check if the reference is on the packed-refs file
	refs, err := b.parsePackedRefs()
	if err != nil {
		return xerrors.Errorf("could not check %s: %w", gitpath.PackedRefsPath, err)
	}
	if _, ok := refs[ref.Name()]; ok {
		return plumbing.ErrRefExists
	}

	return b.WriteReference(ref)
}