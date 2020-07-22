package git

import (
	"errors"
	"path/filepath"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/backend/fsbackend"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
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

// InitRepositoryWithOptions initialize a new git repository by creating
// the .git directory in the given path, which is where almost everything
// that Git stores and manipulates is located.
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

// IsBare returns whether the repo is bare or not.
// A bare repo doesn't have a workign tree
func (r *Repository) IsBare() bool {
	return r.wt == nil
}

// GetObject returns the object matching the given ID
func (r *Repository) GetObject(oid plumbing.Oid) (*object.Object, error) {
	return r.dotGit.Object(oid)
}

// GetCommit returns the commit matching the given SHA
func (r *Repository) GetCommit(oid plumbing.Oid) (*object.Commit, error) {
	o, err := r.dotGit.Object(oid)
	if err != nil {
		return nil, err
	}
	return o.AsCommit()
}

// GetTree returns the tree matching the given SHA
func (r *Repository) GetTree(oid plumbing.Oid) (*object.Tree, error) {
	o, err := r.dotGit.Object(oid)
	if err != nil {
		return nil, err
	}
	return o.AsTree()
}

// NewBlob creates, stores, and returns a new Blob object
func (r *Repository) NewBlob(data []byte) (*object.Blob, error) {
	o := object.New(object.TypeBlob, data)
	if _, err := r.dotGit.WriteObject(o); err != nil {
		return nil, err
	}
	return object.NewBlob(o), nil
}

// NewCommit creates, stores, and returns a new Commit object
// The head of the reference $refname will be updated to this
// new commit.
// An empty refName will create a detached (loose) commit
// If the reference doesn't exists, it will be created
func (r *Repository) NewCommit(refname string, tree *object.Tree, author object.Signature, opts *object.CommitOptions) (*object.Commit, error) {
	// We first validate the parents actually exists
	for _, id := range opts.ParentsID {
		parent, err := r.dotGit.Object(id)
		if err != nil {
			return nil, xerrors.Errorf("could not retrieve parent %s: %w", id.String(), err)
		}
		if parent.Type() != object.TypeCommit {
			return nil, xerrors.Errorf("invalid type for parent %s. got %d, expected %d", id.String(), parent.Type(), parent.Type())
		}
	}

	c := object.NewCommit(tree.ID(), author, opts)
	o := c.ToObject()
	if _, err := r.dotGit.WriteObject(o); err != nil {
		return nil, xerrors.Errorf("could not write the object to the odb: %w", err)
	}

	// If we have a refname then the
	if refname != "" {
		ref := plumbing.NewReference(refname, o.ID())
		if err := r.dotGit.WriteReference(ref); err != nil {
			return nil, xerrors.Errorf("could not update the HEAD of %s: %w", refname, err)
		}
	}

	return o.AsCommit()
}

// NewDetachedCommit creates, stores, and returns a new Commit object
// not attached to any reference
func (r *Repository) NewDetachedCommit(tree *object.Tree, author object.Signature, opts *object.CommitOptions) (*object.Commit, error) {
	return r.NewCommit("", tree, author, opts)
}
