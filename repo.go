package git

import (
	"compress/zlib"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"errors"

	"github.com/Nivl/git-go/internal/readutil"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/Nivl/git-go/plumbing/packfile"
	"golang.org/x/xerrors"
	"gopkg.in/ini.v1"
)

// List of errors returned by the Repository struct
var (
	ErrRepositoryNotExist           = errors.New("repository does not exist")
	ErrRepositoryUnsupportedVersion = errors.New("repository nor supported")
)

// Repository represent a git repository
// A Git repository is the .git/ folder inside a project.
// This repository tracks all changes made to files in your project,
// building a history over time.
// https://blog.axosoft.com/learning-git-repository/
type Repository struct {
	path        string
	projectPath string
}

// NewRepository creates a new Repository instance. The repository
// will neither be created nor loaded. You must call Load() or Init()
// to get a full repository
func NewRepository(projectPath string) *Repository {
	return &Repository{
		path:        filepath.Join(projectPath, DotGitPath),
		projectPath: projectPath,
	}
}

// LoadRepository loads an existing git repository by reading its
// config file, and returns a Repository instance
func LoadRepository(projectPath string) (*Repository, error) {
	r := NewRepository(projectPath)
	err := r.Load()
	return r, err
}

// InitRepository initialize a new git repository by creating the .git
// directory in the given path, which is where almost everything that
// Git stores and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
func InitRepository(projectPath string) (*Repository, error) {
	r := NewRepository(projectPath)
	err := r.Init()
	return r, err
}

// Init initialize a new git repository by creating the .git directory
// in the given path, which is where almost everything that Git stores
// and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
func (r *Repository) Init() error {
	// We don't check if the repo already exist, and just add what's
	// missing to prevent corruption (ie. if we fail creating the repo
	// retrying would always fail because the .git already exist,
	// and Load() would always fail because the repo is invalid)

	// Create all the default folders
	dirs := []string{
		BranchesPath,
		ObjectsPath,
		RefsTagsPath,
		RefsHeadsPath,
		ObjectsInfoPath,
		ObjectsPackPath,
	}
	for _, d := range dirs {
		fullPath := filepath.Join(r.path, d)
		if err := os.MkdirAll(fullPath, 0750); err != nil {
			return xerrors.Errorf("could not create directory %s: %w", d, err)
		}
	}

	// Create the files with the default content
	// (taken from a repo created on github)
	files := []struct {
		path    string
		content []byte
	}{
		{
			path:    DescriptionPath,
			content: []byte("Unnamed repository; edit this file 'description' to name the repository.\n"),
		},
		{
			path:    HEADPath,
			content: []byte("ref: refs/heads/master\n"),
		},
	}

	for _, f := range files {
		fullPath := filepath.Join(r.path, f.path)
		if err := ioutil.WriteFile(fullPath, f.content, 0644); err != nil {
			return xerrors.Errorf("could not create file %s: %w", f, err)
		}
	}

	if err := r.setDefaultCfg(); err != nil {
		return xerrors.Errorf("could not create config file: %w", err)
	}

	return nil
}

// Load loads an existing git repository by reading its config file,
// and returns a Repository instance
func (r *Repository) Load() error {
	// we first make sure the repo exist
	_, err := os.Stat(r.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return xerrors.Errorf("could not check for %s directory: %w", DotGitPath, err)
		}
		return ErrRepositoryNotExist
	}

	// Load the config file
	// https://git-scm.com/docs/git-config
	cfg, err := ini.Load(filepath.Join(r.path, ConfigPath))
	if err != nil {
		return xerrors.Errorf("could not read config file: %w", err)
	}

	// Validate the config
	repoVersion := cfg.Section(cfgCore).Key(cfgCoreFormatVersion).MustInt(0)
	if repoVersion != 0 {
		return ErrRepositoryUnsupportedVersion
	}

	return nil
}

// setDefaultCfg set and persists the default git configuration for
// the repository
func (r *Repository) setDefaultCfg() error {
	cfg := ini.Empty()

	// Core
	core, err := cfg.NewSection(cfgCore)
	if err != nil {
		return xerrors.Errorf("could not create core section: %w", err)
	}
	coreCfg := map[string]string{
		cfgCoreFormatVersion:     "0",
		cfgCoreFileMode:          "true",
		cfgCoreBare:              "false",
		cfgCoreLogAllRefUpdate:   "true",
		cfgCoreIgnoreCase:        "true",
		cfgCorePrecomposeUnicode: "true",
	}
	for k, v := range coreCfg {
		if _, err := core.NewKey(k, v); err != nil {
			return xerrors.Errorf("could not set %s: %w", k, err)
		}
	}
	return cfg.SaveTo(filepath.Join(r.path, ConfigPath))
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
