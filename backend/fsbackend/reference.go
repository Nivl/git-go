package fsbackend

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/errutil"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/spf13/afero"
	"golang.org/x/xerrors"
)

// Reference returns a stored reference from its name
// ErrRefNotFound is returned if the reference doesn't exists
// This method can be called concurrently
func (b *Backend) Reference(name string) (*ginternals.Reference, error) {
	b.refMu.RLock()
	defer b.refMu.RUnlock()

	finder := func(name string) ([]byte, error) {
		data, ok := b.refs[name]
		if !ok {
			return nil, xerrors.Errorf(`ref "%s": %w`, name, ginternals.ErrRefNotFound)
		}
		return data, nil
	}
	return ginternals.ResolveReference(name, finder)
}

// systemPath returns a path from a ref name
// Ex.: On windows refs/heads/master would return refs\heads\master
func (b *Backend) systemPath(name string) string {
	if os.PathSeparator != '/' {
		name = filepath.FromSlash(name)
	}
	return filepath.Join(b.root, name)
}

// loadRefs loads the references in memory
func (b *Backend) loadRefs() (err error) {
	b.refMu.Lock()
	defer b.refMu.Unlock()

	// We first parse the packed-refs file which may or may not exists
	// and may or may not contain outdated information
	// (outdated information will be overwritten once we parse the
	// on-disk references).
	f, err := b.fs.Open(filepath.Join(b.root, gitpath.PackedRefsPath))
	if err != nil && !xerrors.Is(err, os.ErrNotExist) {
		return xerrors.Errorf("could not open %s: %w", gitpath.PackedRefsPath, err)
	}
	// if the file doesn't exist then there's nothing to do
	if err == nil {
		defer errutil.Close(f, &err)

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
				return xerrors.Errorf("could not parse %s, unexpected data line %d: %w", gitpath.PackedRefsPath, i, ginternals.ErrPackedRefInvalid)
			}
			// the name of the ref is its UNIX path
			b.refs[filepath.ToSlash(parts[1])] = []byte(parts[0])
		}

		if sc.Err() != nil {
			return xerrors.Errorf("could not parse %s: %w", gitpath.PackedRefsPath, err)
		}
	}

	// Now we browse all the references on disk
	// TODO(melvin): Do we really want to stop if we cannot parse one file?
	refsPath := filepath.Join(b.root, gitpath.RefsPath)
	err = afero.Walk(b.fs, refsPath, func(path string, info os.FileInfo, e error) error {
		// if refsPath doesn't exists this will return nil and skip the error
		// this is useful in case where the repo is empty and has no
		// references yet
		if path == refsPath {
			return nil
		}

		if e != nil {
			return xerrors.Errorf("could not walk %s: %w", path, e)
		}
		if info.IsDir() {
			return nil
		}
		// TODO(melvin): for security reason we should limit the amount of
		// data we can read
		data, e := afero.ReadFile(b.fs, path)
		if e != nil {
			return xerrors.Errorf("could not read reference at %s: %w", path, e)
		}
		relpath, e := filepath.Rel(b.root, path)
		if e != nil {
			return e // the error message is already pretty descriptive
		}
		// the name of the ref is its UNIX path
		b.refs[filepath.ToSlash(relpath)] = data
		return nil
	})
	if err != nil {
		return xerrors.Errorf("could not browse the refs directory: %w", err)
	}

	// Now we look for the special HEADs references:
	headPaths := []string{
		ginternals.Head,
		ginternals.FetchHead,
		ginternals.OrigHead,
		ginternals.MergeHead,
		ginternals.CherryPickHead,
	}
	for _, path := range headPaths {
		data, err := afero.ReadFile(b.fs, filepath.Join(b.root, path))
		if err != nil {
			if xerrors.Is(err, os.ErrNotExist) {
				continue
			}
			return xerrors.Errorf("could not read reference at %s: %w", path, err)
		}
		b.refs[path] = data
	}

	return nil
}

// WriteReference writes the given reference on disk. If the
// reference already exists it will be overwritten
func (b *Backend) WriteReference(ref *ginternals.Reference) error {
	b.refMu.Lock()
	defer b.refMu.Unlock()
	return b.writeReference(ref)
}

// WriteReferenceSafe writes the given reference on disk.
// ErrRefExists is returned if the reference already exists
func (b *Backend) WriteReferenceSafe(ref *ginternals.Reference) error {
	b.refMu.Lock()
	defer b.refMu.Unlock()

	if _, ok := b.refs[ref.Name()]; ok {
		return ginternals.ErrRefExists
	}

	return b.writeReference(ref)
}

// writeReference writes the given reference on disk. If the
// reference already exists it will be overwritten
func (b *Backend) writeReference(ref *ginternals.Reference) error {
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

	refPath := b.systemPath(ref.Name())
	data := []byte(target)
	err := afero.WriteFile(b.fs, refPath, data, 0o644)
	if err != nil {
		return xerrors.Errorf("could not persist reference to disk: %w", err)
	}
	b.refs[ref.Name()] = data
	return nil
}
