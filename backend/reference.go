package backend

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/errutil"
	"github.com/spf13/afero"
)

// Reference returns a stored reference from its name
// ErrRefNotFound is returned if the reference doesn't exists
// This method can be called concurrently
func (b *Backend) Reference(name string) (*ginternals.Reference, error) {
	finder := func(name string) ([]byte, error) {
		data, ok := b.refs.Load(name)
		if !ok {
			return nil, fmt.Errorf(`ref "%s": %w`, name, ginternals.ErrRefNotFound)
		}
		return data.([]byte), nil
	}
	return ginternals.ResolveReference(name, finder)
}

// systemPath returns a path from a ref name
// Ex.: On windows refs/heads/master would return refs\heads\master
func (b *Backend) systemPath(name string) string {
	name = filepath.FromSlash(name)
	return filepath.Join(b.Path(), name)
}

// loadRefs loads the references in memory
func (b *Backend) loadRefs() (err error) {
	// We first parse the packed-refs file which may or may not exists
	// and may or may not contain outdated information
	// (outdated information will be overwritten once we parse the
	// on-disk references).
	packedRefPath := ginternals.PackedRefsPath(b.config)
	f, err := b.fs.Open(packedRefPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not open %s: %w", packedRefPath, err)
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
				return fmt.Errorf("could not parse %s, unexpected data line %d: %w", packedRefPath, i, ginternals.ErrPackedRefInvalid)
			}
			// the name of the ref is its UNIX path
			b.refs.Store(filepath.ToSlash(parts[1]), []byte(parts[0]))
		}

		if sc.Err() != nil {
			return fmt.Errorf("could not parse %s: %w", packedRefPath, err)
		}
	}

	// Now we browse all the references on disk
	// TODO(melvin): Do we really want to stop if we cannot parse one file?
	refsPath := ginternals.RefsPath(b.config)
	err = afero.Walk(b.fs, refsPath, func(path string, info fs.FileInfo, e error) error {
		// if refsPath doesn't exists this will return nil and skip the error
		// this is useful in case where the repo is empty and has no
		// references yet
		if path == refsPath {
			return nil
		}

		if e != nil {
			return fmt.Errorf("could not walk %s: %w", path, e)
		}
		if info.IsDir() {
			return nil
		}
		// TODO(melvin): for security reason we should limit the amount of
		// data we can read
		data, e := afero.ReadFile(b.fs, path)
		if e != nil {
			return fmt.Errorf("could not read reference at %s: %w", path, e)
		}
		relpath, e := filepath.Rel(b.Path(), path)
		if e != nil {
			return e //nolint:wrapcheck // the error message is already pretty descriptive
		}
		// the name of the ref is its UNIX path
		b.refs.Store(filepath.ToSlash(relpath), data)
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not browse the refs directory: %w", err)
	}

	// Now we look for the special HEADs references:
	headPaths := []string{
		ginternals.Head,
		// TODO(melvin): Removed until we support the format
		// ginternals.FetchHead,
		ginternals.OrigHead,
		ginternals.MergeHead,
		ginternals.CherryPickHead,
	}
	for _, path := range headPaths {
		data, err := afero.ReadFile(b.fs, filepath.Join(b.Path(), path))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("could not read reference at %s: %w", path, err)
		}
		b.refs.Store(path, data)
	}

	return nil
}

// WriteReference writes the given reference on disk. If the
// reference already exists it will be overwritten
func (b *Backend) WriteReference(ref *ginternals.Reference) error {
	return b.writeReference(ref)
}

// WriteReferenceSafe writes the given reference on disk.
// ErrRefExists is returned if the reference already exists
func (b *Backend) WriteReferenceSafe(ref *ginternals.Reference) error {
	if _, ok := b.refs.Load(ref.Name()); ok {
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

	var target string
	switch ref.Type() {
	case ginternals.SymbolicReference:
		target = fmt.Sprintf("ref: %s\n", ref.SymbolicTarget())
	case ginternals.OidReference:
		target = fmt.Sprintf("%s\n", ref.Target().String())
	default:
		return fmt.Errorf("reference type %d: %w", ref.Type(), ginternals.ErrUnknownRefType)
	}

	refPath := b.systemPath(ref.Name())
	// Since we can have `/` in the ref name, we need to create
	// the path on the FS
	dir := filepath.Dir(refPath)
	err := b.fs.MkdirAll(dir, 0o755)
	if err != nil {
		// TODO(melvin): This fails if someone creates a ref
		// named ml/foo and then another ref named ml/foo/bar since
		// foo is a file. We should probably return a better error
		// message in this case (and potentially check this in IsRefNameValid?)
		return fmt.Errorf("could not persist reference to disk: %w", err)
	}
	// We can now create the actual file
	data := []byte(target)
	err = afero.WriteFile(b.fs, refPath, data, 0o644)
	if err != nil {
		return fmt.Errorf("could not persist reference to disk: %w", err)
	}
	b.refs.Store(ref.Name(), data)
	return nil
}

// WalkReferences runs the provided method on all the references
func (b *Backend) WalkReferences(f RefWalkFunc) error {
	var topError error
	b.refs.Range(func(key, value interface{}) bool {
		name, ok := key.(string)
		if !ok {
			//nolint:goerr113 // no need to wrap the error, this would only be caused by a bug in the codebase
			topError = fmt.Errorf("invalid key type for %s. expected string got %T", name, key)
			return false
		}
		ref, err := b.Reference(name)
		if err != nil {
			topError = fmt.Errorf("could not resolve reference %s: %w", name, err)
			return false
		}

		if err = f(ref); err != nil {
			if err != WalkStop { //nolint:errorlint,goerr113 // it's a fake error so no need to use Error.Is()
				topError = err
			}
			return false
		}
		return true
	})

	return topError
}
