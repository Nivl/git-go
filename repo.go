package git

import (
	"compress/zlib"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"errors"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/backend/fsbackend"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/readutil"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/Nivl/git-go/plumbing/packfile"
	"github.com/spf13/afero"
	"golang.org/x/xerrors"
)

// List of errors returned by the Repository struct
var (
	ErrRepositoryNotExist           = errors.New("repository does not exist")
	ErrRepositoryUnsupportedVersion = errors.New("repository nor supported")
	ErrRepositoryExists             = errors.New("repository already exists")
)

// Repository represent a git repository
// A Git repository is the .git/ folder inside a project.
// This repository tracks all changes made to files in your project,
// building a history over time.
// https://blog.axosoft.com/learning-git-repository/
type Repository struct {
	dotGitPath string
	dotGit     backend.Backend
	repoRoot   string
	wt         afero.Fs
}

// InitOptions contains all the optional data used to initialized a
// repository
type InitOptions struct {
	// IsBare represents whether a bare repository will be created or not
	IsBare bool
	// GitBackend represents the underlying backend to use to init the
	// repository and interact with the odb
	// By default the filesystem will be used
	GitBackend backend.Backend
	// WorkingTreeBackend represents the underlying backend to use to
	// interact with the working tree.
	// By default the filesystem will be used
	// Setting this is useless if IsBare is set to true
	WorkingTreeBackend afero.Fs
}

// InitRepository initialize a new git repository by creating the .git
// directory in the given path, which is where almost everything that
// Git stores and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
func InitRepository(repoPath string) (*Repository, error) {
	return InitRepositoryWithOptions(repoPath, InitOptions{})
}

// Init initialize a new git repository by creating the .git directory
// in the given path, which is where almost everything that Git stores
// and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
func InitRepositoryWithOptions(repoPath string, opts InitOptions) (*Repository, error) {
	dotGitPath := repoPath
	if !opts.IsBare {
		dotGitPath = filepath.Join(repoPath, gitpath.DotGitPath)
	}
	r := &Repository{
		repoRoot:   repoPath,
		dotGitPath: dotGitPath,
	}

	if opts.GitBackend == nil {
		r.dotGit = fsbackend.New(r.dotGitPath)
	}

	if !opts.IsBare {
		r.wt = opts.WorkingTreeBackend
		if r.wt == nil {
			r.wt = afero.NewOsFs()
		}
	}

	if err := r.dotGit.Init(); err != nil {
		return nil, err
	}

	ref := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.MasterLocalRef)
	if err := r.dotGit.WriteReference(ref); err != nil {
		if xerrors.Is(err, plumbing.ErrRefExists) {
			return nil, ErrRepositoryExists
		}
		return nil, err
	}

	return r, nil
}

// OpenOptions contains all the optional data used to open a
// repository
type OpenOptions struct {
	// IsBare represents whether a bare repository will be created or not
	IsBare bool
	// GitBackend represents the underlying backend to use to init the
	// repository and interact with the odb
	// By default the filesystem will be used
	GitBackend backend.Backend
	// WorkingTreeBackend represents the underlying backend to use to
	// interact with the working tree.
	// By default the filesystem will be used
	// Setting this is useless if IsBare is set to true
	WorkingTreeBackend afero.Fs
}

// OpenRepository loads an existing git repository by reading its
// config file, and returns a Repository instance
func OpenRepository(repoPath string) (*Repository, error) {
	return OpenRepositoryWithOptions(repoPath, OpenOptions{})
}

// OpenRepositoryWithOptions loads an existing git repository by reading
// its config file, and returns a Repository instance
func OpenRepositoryWithOptions(repoPath string, opts OpenOptions) (*Repository, error) {
	dotGitPath := repoPath
	if !opts.IsBare {
		dotGitPath = filepath.Join(repoPath, gitpath.DotGitPath)
	}
	r := &Repository{
		repoRoot:   repoPath,
		dotGitPath: dotGitPath,
	}

	if opts.GitBackend == nil {
		r.dotGit = fsbackend.New(r.dotGitPath)
	}

	if !opts.IsBare {
		r.wt = opts.WorkingTreeBackend
		if r.wt == nil {
			r.wt = afero.NewOsFs()
		}
	}

	// since we can't check if the directory exists on disk to
	// validate if the repo exists, we're instead going to see if HEAD
	// exists (since it should always be there)
	_, err := r.dotGit.Reference(plumbing.HEAD)
	if err != nil {
		return nil, ErrRepositoryNotExist
	}

	// TODO(melvin): Config check temporarily removed during starage
	// refactor to limit size of PR/sCommits
	// Load the config file
	// https://git-scm.com/docs/git-config
	// cfg, err := ini.Load(filepath.Join(r.path, ConfigPath))
	// if err != nil {
	// 	return xerrors.Errorf("could not read config file: %w", err)
	// }

	// // Validate the config
	// repoVersion := cfg.Section(cfgCore).Key(cfgCoreFormatVersion).MustInt(0)
	// if repoVersion != 0 {
	// 	return ErrRepositoryUnsupportedVersion
	// }

	return r, nil
}

func (r *Repository) IsBare() bool {
	return r.wt == nil
}

func (r *Repository) getLooseObject(oid plumbing.Oid) (*object.Object, error) {
	strOid := oid.String()

	p := r.looseObjectPath(strOid)
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
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

func (r *Repository) getObjectFromPackfile(oid plumbing.Oid) (*object.Object, error) {
	// Not found? Let's find packfiles
	p := filepath.Join(r.path, ObjectsPackPath)
	packfiles := []string{}
	err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// fmt.Printf("cannot read file %s: %s\n", info.Name(), err.Error())
			// TODO(melvin): log error
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
		pf, err := packfile.NewFromFile(r, packFilePath)
		if err != nil {
			return nil, xerrors.Errorf("could not open packfile: %w", err)
		}
		do, err := pf.GetObject(oid)
		if err == nil {
			return do, nil
		}
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			continue
		}
		return nil, err
	}
	return nil, plumbing.ErrObjectNotFound
}

// GetObject returns the object matching the given SHA
// The format of an object is an ascii encoded type, an ascii encoded
// space, then an ascii encoded length of the object, then a null
// character, then the body of the object
func (r *Repository) GetObject(oid plumbing.Oid) (*object.Object, error) {
	// First let's look for loose objects
	o, err := r.getLooseObject(oid)
	if err == nil {
		return o, nil
	}
	if !os.IsNotExist(err) {
		return nil, xerrors.Errorf("failed looking for loose object: %w", err)
	}

	o, err = r.getObjectFromPackfile(oid)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// WriteObject writes an object on disk and return its Oid
func (r *Repository) WriteObject(o *object.Object) (plumbing.Oid, error) {
	data, err := o.Compress()
	if err != nil {
		return plumbing.NullOid, xerrors.Errorf("unsupported object type %s", o.Type())
	}

	// Persist the data on disk
	sha := o.ID.String()
	p := r.looseObjectPath(sha)
	if err = ioutil.WriteFile(p, data, 0644); err != nil {
		return plumbing.NullOid, xerrors.Errorf("could not persist object %s at path %s: %w", sha, p, err)
	}

	return o.ID, nil
}

// looseObjectPath returns the absolute path of an object
// .git/object/first_2_chars_of_sha/remaining_chars_of_sha
// Ex. path of fcfe68a0e44e04bd7fd564fc0b75f1ae457e18b3 is:
// .git/objects/fc/fe68a0e44e04bd7fd564fc0b75f1ae457e18b3
func (r *Repository) looseObjectPath(sha string) string {
	return filepath.Join(r.path, ObjectsPath, sha[:2], sha[2:])
}

// NewBlob creates, stores, and returns a new Blob object
func (r *Repository) NewBlob(data []byte) (*object.Blob, error) {
	o := object.New(object.TypeBlob, data)
	_, err := o.Compress()
	if err != nil {
		return nil, xerrors.Errorf("could not compress object: %w", err)
	}
	// TODO(melvin): actually store the data
	return object.NewBlob(o), nil
}
