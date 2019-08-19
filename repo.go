package git

import (
	"compress/zlib"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
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
	return r, r.Load()
}

// InitRepository initialize a new git repository by creating the .git
// directory in the given path, which is where almost everything that
// Git stores and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
func InitRepository(projectPath string) (*Repository, error) {
	r := NewRepository(projectPath)
	return r, r.Init()
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
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return errors.Wrapf(err, "could not create directory %s", d)
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
			return errors.Wrapf(err, "could not create file %s", f)
		}
	}

	if err := r.setDefaultCfg(); err != nil {
		return errors.Wrap(err, "could not create config file")
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
			return errors.Wrapf(err, "could not check for %s directory", DotGitPath)
		}
		return ErrRepositoryNotExist
	}

	// Load the config file
	// https://git-scm.com/docs/git-config
	cfg, err := ini.Load(filepath.Join(r.path, ConfigPath))
	if err != nil {
		return errors.Wrapf(err, "could not read config file")
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
		return errors.Wrap(err, "could not create core section")
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
			return errors.Wrapf(err, "could not set %s", k)
		}
	}
	return cfg.SaveTo(filepath.Join(r.path, ConfigPath))
}

func (r *Repository) getDanglingObject(oid Oid) (*Object, error) {
	strOid := oid.String()

	p := r.danglingObjectPath(strOid)
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, errors.Wrapf(err, "could not find object %s at path %s", strOid, p)
	}
	defer f.Close()

	// Objects are zlib encoded
	zlibReader, err := zlib.NewReader(f)
	if err != nil {
		return nil, errors.Wrapf(err, "could not decompress parts of object %s at path %s", strOid, p)
	}
	defer zlibReader.Close()

	// We directly read the entire file since most of it is the content we
	// need, this allows us to be able to easily store the object's content
	buff, err := ioutil.ReadAll(zlibReader)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read object %s at path %s", strOid, p)
	}

	o := &Object{
		ID: oid,
	}
	// we keep track of where we're at in the buffer
	pointerPos := 0

	// the type of the object starts at offset 0 and ends a the first
	// space character that we'll need to trim
	typ := readTo(buff, ' ')
	// typ, err := objectData.ReadString(' ')
	if typ == nil {
		return nil, errors.Wrapf(err, "could not find object type for %s at path %s", strOid, p)
	}

	o.typ, err = NewObjectTypeFromString(string(typ))
	if err != nil {
		return nil, errors.Errorf("unsupported type %s for object %s at path %s", string(typ), strOid, p)
	}
	pointerPos += len(typ)
	pointerPos++ // one more for the space

	// The size of the object starts after the space and ends at a NULL char
	// That we'll need to trim.
	// A NULL char is represented by 0 (dec), 000 (octal), or 0x00 (hex)
	// type "man ascii" in a terminal for more information
	size := readTo(buff[pointerPos:], 0)
	if size == nil {
		return nil, errors.Wrapf(err, "could not find object size for %s at path %s", strOid, p)
	}
	o.size, err = strconv.Atoi(string(size))
	if err != nil {
		return nil, errors.Wrapf(err, "invalid size %s for object %s at path %s", size, strOid, p)
	}
	pointerPos += len(size)
	pointerPos++ // one more for the NULL char
	o.content = buff[pointerPos:]

	return o, nil
}

func (r *Repository) getObjectFromPackfile(oid Oid) (*Object, error) {
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
		if filepath.Ext(info.Name()) != ExtPackfile {
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
		pf, err := NewPackFromFile(r, packFilePath)
		if err != nil {
			return nil, errors.Wrap(err, "could not open packfile")
		}
		do, err := pf.GetObject(oid)
		if err == nil {
			return do, nil
		}
		if err == ErrObjectNotFound {
			continue
		}
		return nil, err
	}
	return nil, ErrObjectNotFound
}

// GetObject returns the object matching the given SHA
// The format of an object is an ascii encoded type, an ascii encoded
// space, then an ascii encoded length of the object, then a null
// character, then the body of the object
func (r *Repository) GetObject(oid Oid) (*Object, error) {
	// First let's look for dangling objects
	o, err := r.getDanglingObject(oid)
	if err == nil {
		return o, nil
	}
	if !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "failed looking for danglin object")
	}

	o, err = r.getObjectFromPackfile(oid)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// WriteObject writes an object on disk and return its Oid
func (r *Repository) WriteObject(o *Object) (Oid, error) {
	oid, data, err := o.Compress()
	if err != nil {
		return NullOid, errors.Errorf("unsupported object type %s", o.Type())
	}

	// Persist the data on disk
	sha := oid.String()
	p := r.danglingObjectPath(sha)
	if err = ioutil.WriteFile(p, data, 0644); err != nil {
		return NullOid, errors.Wrapf(err, "could not persist object %s at path %s", sha, p)
	}

	return oid, nil
}

// danglingObjectPath returns the absolute path of an object
// .git/object/first_2_chars_of_sha/remaining_chars_of_sha
// Ex. path of fcfe68a0e44e04bd7fd564fc0b75f1ae457e18b3 is:
// .git/objects/fc/fe68a0e44e04bd7fd564fc0b75f1ae457e18b3
func (r *Repository) danglingObjectPath(sha string) string {
	return filepath.Join(r.path, ObjectsPath, sha[:2], sha[2:])
}
