package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/spf13/afero"
)

// List of errors returned by the Repository struct
var (
	ErrRepositoryNotExist           = errors.New("repository does not exist")
	ErrRepositoryUnsupportedVersion = errors.New("repository nor supported")
	ErrTagNotFound                  = errors.New("tag not found")
	ErrTagExists                    = errors.New("tag already exists")
	ErrNotADirectory                = errors.New("not a directory")
	ErrInvalidBranchName            = errors.New("invalid branch name")
)

// Repository represent a git repository
// A Git repository is the .git/ folder inside a project.
// This repository tracks all changes made to files in your project,
// building a history over time.
// https://blog.axosoft.com/learning-git-repository/
type Repository struct {
	Config   *config.Config
	workTree afero.Fs
	dotGit   *backend.Backend

	shouldCleanBackend bool
}

// InitOptions contains all the optional data used to initialized a
// repository
type InitOptions struct {
	// GitBackend represents the underlying backend to use to init the
	// repository and interact with the odb
	// By default the filesystem will be used
	GitBackend *backend.Backend
	// WorkingTreeBackend represents the underlying backend to use to
	// interact with the working tree.
	// By default the filesystem will be used
	// Setting this is useless if IsBare is set to true
	WorkingTreeBackend afero.Fs
	// InitialBranchName represents the name of the default branch to use
	// Defaults to master
	InitialBranchName string
	// IsBare represents whether a bare repository will be created or not
	IsBare bool
	// Symlink will create a .git text file in the working tree that points
	// toward the actual repository
	Symlink bool
}

// InitRepository initialize a new git repository by creating the .git
// directory in the given path, which is where almost everything that
// Git stores and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
//
// This assumes:
// - The repo is not bare (see WithOptions)
// - We're not interested in env vars (see WithParams)
// - The git dir is in the working tree under .git
func InitRepository(workTreePath string) (*Repository, error) {
	return InitRepositoryWithOptions(workTreePath, InitOptions{})
}

// InitRepositoryWithOptions initialize a new git repository by creating
// the .git directory in the given path, which is where almost everything
// that Git stores and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
//
// This assumes:
// - We're not interested in env vars (see WithParams)
// - The git dir is in the working tree under .git
func InitRepositoryWithOptions(rootPath string, opts InitOptions) (r *Repository, err error) {
	WorkTreePath := rootPath
	GitDirPath := filepath.Join(rootPath, config.DefaultDotGitDirName)
	if opts.IsBare {
		WorkTreePath = ""
		GitDirPath = rootPath
	}

	params, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
		WorkTreePath: WorkTreePath,
		GitDirPath:   GitDirPath,
		IsBare:       opts.IsBare,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get the repo params: %w", err)
	}
	return InitRepositoryWithParams(params, opts)
}

// InitRepositoryWithParams initialize a new git repository by creating the .git
// directory in the given path, which is where almost everything that
// Git stores and manipulates is located.
// https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain#ch10-git-internals
//
// This method makes no assumptions
func InitRepositoryWithParams(cfg *config.Config, opts InitOptions) (r *Repository, err error) {
	r = &Repository{
		Config: cfg,
	}

	// Validate the branch name
	branchName := opts.InitialBranchName
	if branchName == "" {
		branchName, _ = cfg.FromFile().DefaultBranch()
		if branchName == "" {
			branchName = ginternals.Master
		}
	}
	if !ginternals.IsRefNameValid(branchName) {
		return nil, ErrInvalidBranchName
	}

	// if the repo is not bare, then we need to make sure to create
	// the working tree
	if !opts.IsBare {
		info, err := os.Stat(cfg.WorkTreePath)
		switch err { //nolint:errorlint // we only want nil or not nil
		case nil:
			if !info.IsDir() {
				return nil, fmt.Errorf("invalid path: %w", ErrNotADirectory)
			}
		default:
			if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("could not check %s: %w", cfg.WorkTreePath, err)
			}
			err = os.MkdirAll(cfg.WorkTreePath, 0o755)
			if err != nil {
				return nil, fmt.Errorf("could not create %s: %w", cfg.WorkTreePath, err)
			}
		}

		r.workTree = opts.WorkingTreeBackend
		if r.workTree == nil {
			r.workTree = afero.NewOsFs()
		}
	}

	if opts.GitBackend == nil {
		r.dotGit, err = backend.NewFS(cfg)
		if err != nil {
			return nil, fmt.Errorf("could not create backend: %w", err)
		}
		r.shouldCleanBackend = true
		// we pass the repo by copy because in case of error the pointer
		// will be changed to nil
		defer func(r *Repository) {
			if err != nil {
				r.dotGit.Close() //nolint:errcheck // it already failed
			}
		}(r)
	}

	err = r.dotGit.InitWithOptions(branchName, backend.InitOptions{
		CreateSymlink: opts.Symlink,
	})
	if err != nil {
		return nil, err
	}

	return r, err
}

// OpenOptions contains all the optional data used to open a
// repository
type OpenOptions struct {
	// GitBackend represents the underlying backend to use to init the
	// repository and interact with the odb
	// By default the filesystem will be used
	GitBackend *backend.Backend
	// WorkingTreeBackend represents the underlying backend to use to
	// interact with the working tree.
	// By default the filesystem will be used
	// Setting this is useless if IsBare is set to true
	WorkingTreeBackend afero.Fs
	// GitDirPath represents the path to the .git directory
	// Defaults to .git
	// IsBare represents whether a bare repository will be created or not
	IsBare bool
}

// OpenRepository loads an existing git repository by reading its
// config file, and returns a Repository instance
//
// This assumes:
// - The repo is not bare (see WithOptions)
// - We're not interested in env vars (see WithParams)
// - The git dir is in the working tree under .git
func OpenRepository(workTreePath string) (*Repository, error) {
	return OpenRepositoryWithOptions(workTreePath, OpenOptions{})
}

// OpenRepositoryWithOptions loads an existing git repository by reading
// its config file, and returns a Repository instance
//
// This assumes:
// - We're not interested in env vars (see WithParams)
// - The git dir is in the working tree under .git
func OpenRepositoryWithOptions(rootPath string, opts OpenOptions) (r *Repository, err error) {
	WorkTreePath := rootPath
	GitDirPath := filepath.Join(rootPath, config.DefaultDotGitDirName)
	if opts.IsBare {
		WorkTreePath = ""
		GitDirPath = rootPath
	}

	params, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
		WorkTreePath: WorkTreePath,
		GitDirPath:   GitDirPath,
		IsBare:       opts.IsBare,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get the repo params: %w", err)
	}
	return OpenRepositoryWithParams(params, opts)
}

// OpenRepositoryWithParams loads an existing git repository by reading
// its config file, and returns a Repository instance
//
// This method makes no assumptions
func OpenRepositoryWithParams(cfg *config.Config, opts OpenOptions) (r *Repository, err error) {
	r = &Repository{
		Config: cfg,
	}

	if !opts.IsBare {
		r.workTree = opts.WorkingTreeBackend
		if r.workTree == nil {
			r.workTree = afero.NewOsFs()
		}
	}

	if opts.GitBackend == nil {
		r.dotGit, err = backend.NewFS(cfg)
		if err != nil {
			return nil, fmt.Errorf("could not create backend: %w", err)
		}
		r.shouldCleanBackend = true
		// we pass the repo by copy because in case of error the pointer
		// will be chaged to nil
		defer func(r *Repository) {
			if err != nil {
				r.dotGit.Close() //nolint:errcheck // it already failed
			}
		}(r)
	}

	// since we can't check if the directory exists on disk to
	// validate if the repo exists, we're instead going to see if HEAD
	// exists (since it should always be there)
	_, err = r.dotGit.Reference(ginternals.Head)
	if err != nil {
		return nil, ErrRepositoryNotExist
	}

	return r, nil
}

// IsBare returns whether the repo is bare or not.
// A bare repo doesn't have a workign tree
func (r *Repository) IsBare() bool {
	return r.workTree == nil
}

// GetObject returns the object matching the given ID
func (r *Repository) GetObject(oid ginternals.Oid) (*object.Object, error) {
	return r.dotGit.Object(oid)
}

// GetCommit returns the commit matching the given SHA
func (r *Repository) GetCommit(oid ginternals.Oid) (*object.Commit, error) {
	o, err := r.dotGit.Object(oid)
	if err != nil {
		return nil, fmt.Errorf("could not get object: %w", err)
	}
	return o.AsCommit()
}

// GetTree returns the tree matching the given SHA
func (r *Repository) GetTree(oid ginternals.Oid) (*object.Tree, error) {
	o, err := r.dotGit.Object(oid)
	if err != nil {
		return nil, fmt.Errorf("could not get object: %w", err)
	}
	return o.AsTree()
}

// GetTag returns the reference for the given tag
// To know if the tag is annoted or lightweight, call repo.GetObject()
// on the reference's target ad make sure that the returned object is
// not a tag with the same name (note that it's technically possible for
// a tag to target another tag)
func (r *Repository) GetTag(name string) (*ginternals.Reference, error) {
	ref, err := r.dotGit.Reference(ginternals.LocalTagFullName(name))
	if err != nil {
		return nil, ErrTagNotFound
	}
	return ref, nil
}

// GetReference returns the reference matching the given name
func (r *Repository) GetReference(name string) (*ginternals.Reference, error) {
	return r.dotGit.Reference(name)
}

// NewBlob creates, stores, and returns a new Blob object
func (r *Repository) NewBlob(data []byte) (*object.Blob, error) {
	o := object.New(object.TypeBlob, data)
	if _, err := r.dotGit.WriteObject(o); err != nil {
		return nil, fmt.Errorf("could not write object: %w", err)
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
			return nil, fmt.Errorf("could not retrieve parent %s: %w", id.String(), err)
		}
		if parent.Type() != object.TypeCommit {
			return nil, fmt.Errorf("invalid type for parent %s. got %d, expected %d: %w", id.String(), parent.Type(), parent.Type(), object.ErrObjectInvalid)
		}
	}

	c := object.NewCommit(tree.ID(), author, opts)
	o := c.ToObject()
	if _, err := r.dotGit.WriteObject(o); err != nil {
		return nil, fmt.Errorf("could not write the object to the odb: %w", err)
	}

	// If we have a refname then we update it
	if refname != "" {
		ref := ginternals.NewReference(refname, o.ID())
		if err := r.dotGit.WriteReference(ref); err != nil {
			return nil, fmt.Errorf("could not update the HEAD of %s: %w", refname, err)
		}
	}

	return o.AsCommit()
}

// NewDetachedCommit creates, stores, and returns a new Commit object
// not attached to any reference
func (r *Repository) NewDetachedCommit(tree *object.Tree, author object.Signature, opts *object.CommitOptions) (*object.Commit, error) {
	return r.NewCommit("", tree, author, opts)
}

// NewTag creates, stores, and returns a new annoted tag
func (r *Repository) NewTag(p *object.TagParams) (*object.Tag, error) {
	found, err := r.dotGit.HasObject(p.Target.ID())
	if err != nil {
		return nil, fmt.Errorf("could not check if target exists: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("target doesn't exists: %w", object.ErrObjectInvalid)
	}

	// We first make sure the tag doesn't already exist
	refname := ginternals.LocalTagFullName(p.Name)
	_, err = r.dotGit.Reference(refname)
	if err == nil {
		return nil, ErrTagExists
	}
	if !errors.Is(err, ginternals.ErrRefNotFound) {
		return nil, fmt.Errorf("could not check if tag already exists: %w", err)
	}

	// We create the tag and persist it to the object database
	o := object.NewTag(p).ToObject()
	if _, err := r.dotGit.WriteObject(o); err != nil {
		return nil, fmt.Errorf("could not write the object to the odb: %w", err)
	}

	// We create the reference for the tag
	ref := ginternals.NewReference(refname, o.ID())
	if err := r.dotGit.WriteReference(ref); err != nil {
		return nil, fmt.Errorf("could not write the ref at %s: %w", refname, err)
	}

	return o.AsTag()
}

// NewLightweightTag creates, stores, and returns a lightweight tag
func (r *Repository) NewLightweightTag(tag string, targetID ginternals.Oid) (*ginternals.Reference, error) {
	// let's make sure the object exists
	found, err := r.dotGit.HasObject(targetID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve targeted object: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("target : %w", object.ErrObjectInvalid)
	}

	refname := ginternals.LocalTagFullName(tag)
	_, err = r.dotGit.Reference(refname)
	if err == nil {
		return nil, ErrTagExists
	}
	if !errors.Is(err, ginternals.ErrRefNotFound) {
		return nil, fmt.Errorf("could not check if tag already exists: %w", err)
	}

	ref := ginternals.NewReference(refname, targetID)
	if err := r.dotGit.WriteReference(ref); err != nil {
		return nil, fmt.Errorf("could not write the ref at %s: %w", refname, err)
	}
	return ref, nil
}

// Close frees the resources used by the repository
func (r *Repository) Close() error {
	if r.shouldCleanBackend {
		return r.dotGit.Close()
	}
	return nil
}
