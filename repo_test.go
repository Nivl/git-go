package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/config"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Parallel()

	t.Run("repo with working tree", func(t *testing.T) {
		t.Parallel()

		// Setup
		d, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		// Run logic
		r, err := InitRepository(d)
		require.NoError(t, err, "failed creating a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		// assert returned repository
		assert.Equal(t, d, r.Config.WorkTreePath)
		assert.Equal(t, ginternals.DotGitPath(r.Config), r.dotGit.Path())
		assert.NotNil(t, r.workTree)
		assert.False(t, r.IsBare(), "repos should not be bare")

		// Default branch should be master
		data, err := os.ReadFile(filepath.Join(ginternals.DotGitPath(r.Config), ginternals.Head))
		require.NoError(t, err)
		require.Equal(t, "ref: refs/heads/master\n", string(data))
	})

	t.Run("bare repo", func(t *testing.T) {
		t.Parallel()

		// Setup
		d, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		// Run logic
		r, err := InitRepositoryWithOptions(d, InitOptions{
			IsBare: true,
		})
		require.NoError(t, err, "failed creating a repo")

		// assert returned repository
		require.Empty(t, r.Config.WorkTreePath)
		require.Equal(t, d, r.dotGit.Path())
		assert.Nil(t, r.workTree)
		assert.True(t, r.IsBare(), "repos should be bare")
	})

	t.Run("repo with a custom .git", func(t *testing.T) {
		t.Parallel()

		d, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		opts, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
			WorkTreePath: d,
			GitDirPath:   filepath.Join(d, "dot-git"),
		})
		require.NoError(t, err)

		// Run logic
		r, err := InitRepositoryWithParams(opts, InitOptions{})
		require.NoError(t, err, "failed creating a repo")

		// assert returned repository
		require.Equal(t, filepath.Join(d, "dot-git"), r.dotGit.Path())
	})

	t.Run("repo with a custom .git and .git/objects", func(t *testing.T) {
		t.Parallel()

		d, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		// Run logic
		e := env.NewFromKVList([]string{
			"GIT_DIR=" + filepath.Join(d, "dot-git"),
			"GIT_OBJECT_DIRECTORY=" + filepath.Join(d, "dot-git-objects"),
		})
		p, err := config.LoadConfig(e, config.LoadConfigOptions{
			IsBare: true,
		})
		require.NoError(t, err)

		// Run logic
		r, err := InitRepositoryWithParams(p, InitOptions{
			IsBare: true,
		})
		require.NoError(t, err, "failed creating a repo")

		// assert returned repository
		require.Equal(t, filepath.Join(d, "dot-git"), r.dotGit.Path())
		require.Equal(t, filepath.Join(d, "dot-git-objects"), ginternals.ObjectsPath(r.Config))
	})

	t.Run("should fail with a path that points to a file", func(t *testing.T) {
		t.Parallel()

		// Setup
		f, cleanup := testutil.TempFile(t)
		t.Cleanup(cleanup)

		// Run logic
		_, err := InitRepositoryWithOptions(f.Name(), InitOptions{
			IsBare: true,
		})
		require.Error(t, err)
		switch runtime.GOOS {
		case "windows":
			require.Contains(t, err.Error(), "The system cannot find the path specified")
		default:
			require.Contains(t, err.Error(), "not a directory")
		}
	})

	t.Run("should NOT fail with a repo that already exists", func(t *testing.T) {
		t.Parallel()

		// Setup
		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		// Run logic
		r, err := InitRepository(repoPath)
		require.NoError(t, err)
		require.NoError(t, r.Close())
	})

	// Windows deals with permission differently
	if runtime.GOOS != "windows" {
		t.Run("should fail creating a repo in a protected directory", func(t *testing.T) {
			t.Parallel()

			dir, cleanup := testutil.TempDir(t)
			t.Cleanup(cleanup)

			target := filepath.Join(dir, "protected")
			err := os.MkdirAll(target, 0o100)
			require.NoError(t, err)

			opts, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
				WorkingDirectory: filepath.Join(target, "wt"),
				SkipGitDirLookUp: true,
			})
			require.NoError(t, err)

			// Run logic
			_, err = InitRepositoryWithParams(opts, InitOptions{})
			require.Error(t, err)
			require.Contains(t, err.Error(), "could not create")
		})
	}

	t.Run("should use provided branch", func(t *testing.T) {
		t.Parallel()

		// Setup
		d, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		// Run logic
		r, err := InitRepositoryWithOptions(d, InitOptions{
			InitialBranchName: "main",
		})
		require.NoError(t, err, "failed creating a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		// assert returned repository
		assert.Equal(t, d, r.Config.WorkTreePath)
		assert.Equal(t, ginternals.DotGitPath(r.Config), r.dotGit.Path())
		assert.NotNil(t, r.workTree)
		assert.False(t, r.IsBare(), "repos should not be bare")

		// Default branch should be master
		data, err := os.ReadFile(filepath.Join(ginternals.DotGitPath(r.Config), ginternals.Head))
		require.NoError(t, err)
		require.Equal(t, "ref: refs/heads/main\n", string(data))
	})

	t.Run("should use provided branch", func(t *testing.T) {
		t.Parallel()

		// Setup
		d, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		// Run logic
		_, err := InitRepositoryWithOptions(d, InitOptions{
			InitialBranchName: "../master",
		})
		require.Error(t, err)
		require.Equal(t, err, ErrInvalidBranchName)
	})

	t.Run("Bare repo", func(t *testing.T) {
		t.Parallel()

		// Setup
		d, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		// Run logic
		r, err := InitRepositoryWithOptions(d, InitOptions{
			IsBare: true,
		})
		require.NoError(t, err, "failed creating a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})
		assert.NoFileExists(t, filepath.Join(d, config.DefaultDotGitDirName))
		assert.FileExists(t, filepath.Join(d, ginternals.Head))
	})
}

func TestOpen(t *testing.T) {
	t.Parallel()

	t.Run("repo with working tree", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close())
		})

		// assert returned repository
		assert.Equal(t, repoPath, r.Config.WorkTreePath)
		assert.Equal(t, ginternals.DotGitPath(r.Config), r.dotGit.Path())
		assert.NotNil(t, r.workTree)
		assert.False(t, r.IsBare(), "repos should not be bare")
	})

	t.Run("bare repo", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)
		repoPath = filepath.Join(repoPath, ".git")

		r, err := OpenRepositoryWithOptions(repoPath, OpenOptions{
			IsBare: true,
		})
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		// assert returned repository
		require.Empty(t, r.Config.WorkTreePath)
		require.Equal(t, repoPath, r.dotGit.Path())
		assert.Nil(t, r.workTree)
		assert.True(t, r.IsBare(), "repos should be bare")
	})

	t.Run("repo with a custom .git", func(t *testing.T) {
		t.Parallel()

		d, cleanupWt := testutil.TempDir(t)
		t.Cleanup(cleanupWt)

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)
		repoPath = filepath.Join(repoPath, ".git")

		p, err := config.LoadConfigSkipEnv(config.LoadConfigOptions{
			WorkTreePath: d,
			GitDirPath:   repoPath,
		})
		require.NoError(t, err)

		// Run logic
		r, err := OpenRepositoryWithParams(p, OpenOptions{})
		require.NoError(t, err, "failed creating a repo")
		require.NoError(t, r.Close())

		// assert returned repository
		require.Equal(t, repoPath, r.dotGit.Path())
	})

	t.Run("should fail if repo doesn't exist", func(t *testing.T) {
		t.Parallel()

		d, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		_, err := OpenRepository(d)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrRepositoryNotExist)
	})

	t.Run("should fail if directory doesn't exist", func(t *testing.T) {
		t.Parallel()

		d, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		_, err := OpenRepository(filepath.Join(d, "404"))
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrRepositoryNotExist)
	})
}

func TestRepositoryObject(t *testing.T) {
	t.Parallel()

	t.Run("loose object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("b07e28976ac8972715598f390964d53cf4dbc1bd")
		require.NoError(t, err)

		obj, err := r.Object(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID())
		assert.Equal(t, object.TypeBlob, obj.Type())
		assert.Equal(t, "package packfile", string(obj.Bytes()[:16]))
	})

	t.Run("Object from packfile", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		obj, err := r.Object(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID())
		assert.Equal(t, object.TypeCommit, obj.Type())
	})
}

func TestRepositoryBlob(t *testing.T) {
	t.Parallel()

	t.Run("should succeed with a valid", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("b07e28976ac8972715598f390964d53cf4dbc1bd")
		require.NoError(t, err)

		obj, err := r.Blob(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID())
		assert.Equal(t, "package packfile", string(obj.Bytes()[:16]))
	})

	t.Run("should fail returning an invalid object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("e16f38976ac8972715598f390964d53cf4dbc1bd")
		require.NoError(t, err)

		_, err = r.Blob(oid)
		require.Error(t, err)
	})
}

func TestRepositoryNewBlob(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
	t.Cleanup(cleanup)

	r, err := OpenRepository(repoPath)
	require.NoError(t, err, "failed loading a repo")
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	data := "abcdefghijklmnopqrstuvwxyz"
	blob, err := r.NewBlob([]byte(data))
	require.NoError(t, err)
	assert.NotEqual(t, ginternals.NullOid, blob.ID())
	assert.Equal(t, []byte(data), blob.Bytes())

	// make sure the blob was persisted
	p := ginternals.LooseObjectPath(r.Config, blob.ID().String())
	assert.FileExists(t, p)
}

func TestRepositoryCommit(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
	t.Cleanup(cleanup)

	r, err := OpenRepository(repoPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	commitOid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
	require.NoError(t, err)

	c, err := r.Commit(commitOid)
	require.NoError(t, err)

	assert.Equal(t, commitOid, c.ID())
	assert.Equal(t, "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3", c.TreeID().String())
	require.Len(t, c.ParentIDs(), 1)
	assert.Equal(t, "6097a04b7a327c4be68f222ca66e61b8e1abe5c1", c.ParentIDs()[0].String())
}

func TestRepositoryReference(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
	t.Cleanup(cleanup)
	r, err := OpenRepository(repoPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	testCases := []struct {
		desc           string
		refName        string
		expectedError  error
		expectedTarget string
	}{
		{
			desc:           "HEAD should work",
			refName:        "HEAD",
			expectedTarget: "bbb720a96e4c29b9950a4c577c98470a4d5dd089",
		},
		{
			desc:           "refs/heads/ml/packfile/tests should work",
			refName:        "refs/heads/ml/packfile/tests",
			expectedTarget: "bbb720a96e4c29b9950a4c577c98470a4d5dd089",
		},
		{
			desc:          "an invalid name should fail",
			refName:       "nope",
			expectedError: ginternals.ErrRefNotFound,
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			ref, err := r.Reference(tc.refName)

			if tc.expectedError != nil {
				assert.True(t, errors.Is(err, tc.expectedError), "wrong error returned")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedTarget, ref.Target().String())
		})
	}
}

func TestRepositoryTree(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
	t.Cleanup(cleanup)

	r, err := OpenRepository(repoPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	treeOid, err := ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
	require.NoError(t, err)

	tree, err := r.Tree(treeOid)
	require.NoError(t, err)

	assert.Equal(t, treeOid, tree.ID())
	require.Len(t, tree.Entries(), 13)
}

func TestRepositoryNewCommit(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(ginternals.LocalBranchFullName(ginternals.Master))
		require.NoError(t, err)

		headCommit, err := r.Commit(ref.Target())
		require.NoError(t, err)

		headTree, err := r.Tree(headCommit.TreeID())
		require.NoError(t, err)

		sig := object.NewSignature("author", "author@domain.tld")
		c, err := r.NewCommit(ginternals.LocalBranchFullName(ginternals.Master), headTree, sig, &object.CommitOptions{
			ParentsID: []ginternals.Oid{headCommit.ID()},
			Message:   "new commit that doesn't do anything",
		})
		require.NoError(t, err)

		// The commit should be findable
		_, err = r.Commit(c.ID())
		require.NoError(t, err)

		// We update the ref since it should have changed
		ref, err = r.dotGit.Reference(ginternals.LocalBranchFullName(ginternals.Master))
		require.NoError(t, err)
		assert.Equal(t, c.ID(), ref.Target())
	})

	t.Run("should fail if a parent is not a commit", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(ginternals.LocalBranchFullName(ginternals.Master))
		require.NoError(t, err)

		headCommit, err := r.Commit(ref.Target())
		require.NoError(t, err)

		headTree, err := r.Tree(headCommit.TreeID())
		require.NoError(t, err)

		sig := object.NewSignature("author", "author@domain.tld")
		_, err = r.NewCommit(ginternals.LocalBranchFullName(ginternals.Master), headTree, sig, &object.CommitOptions{
			ParentsID: []ginternals.Oid{headTree.ID()},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid type for parent")
	})
}

func TestRepositoryNewDetachedCommit(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
	t.Cleanup(cleanup)

	r, err := OpenRepository(repoPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	ref, err := r.dotGit.Reference(ginternals.LocalBranchFullName(ginternals.Master))
	require.NoError(t, err)

	headCommit, err := r.Commit(ref.Target())
	require.NoError(t, err)

	headTree, err := r.Tree(headCommit.TreeID())
	require.NoError(t, err)

	sig := object.NewSignature("author", "author@domain.tld")
	c, err := r.NewDetachedCommit(headTree, sig, &object.CommitOptions{
		ParentsID: []ginternals.Oid{headCommit.ID()},
		Message:   "new commit that doesn't do anything",
	})
	require.NoError(t, err)

	// The commit should be findable
	_, err = r.Commit(c.ID())
	require.NoError(t, err)

	// We update the ref to make sure it's not updated
	updateddRef, err := r.dotGit.Reference(ginternals.LocalBranchFullName(ginternals.Master))
	require.NoError(t, err)
	assert.Equal(t, ref.Target(), updateddRef.Target())
}

func TestRepositoryTag(t *testing.T) {
	t.Parallel()

	t.Run("annotated", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		tagID, err := ginternals.NewOidFromStr("80316e01dbfdf5c2a8a20de66c747ecd4c4bd442")
		require.NoError(t, err)

		tagRef, err := r.Tag("annotated")
		require.NoError(t, err)

		require.Equal(t, tagID, tagRef.Target())

		rawTag, err := r.Object(tagRef.Target())
		require.NoError(t, err)
		tag, err := rawTag.AsTag()
		require.NoError(t, err)

		targettedCommitID, err := ginternals.NewOidFromStr("6097a04b7a327c4be68f222ca66e61b8e1abe5c1")
		require.NoError(t, err)

		assert.Equal(t, tagID, tag.ID())
		assert.Equal(t, "annotated", tag.Name())
		assert.Equal(t, targettedCommitID, tag.Target())
	})

	t.Run("lightweight", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		targettedCommitID, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		tagRef, err := r.Tag("lightweight")
		require.NoError(t, err)

		require.Equal(t, targettedCommitID, tagRef.Target())

		commit, err := r.Commit(tagRef.Target())
		require.NoError(t, err)

		assert.Equal(t, targettedCommitID, commit.ID())
	})

	t.Run("unexisting tag", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		_, err = r.Tag("does-not-exist")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrTagNotFound), "invalid error type")
	})
}

func TestRepositoryNewTag(t *testing.T) {
	t.Parallel()

	t.Run("create a new valid tag", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(ginternals.LocalBranchFullName(ginternals.Master))
		require.NoError(t, err)

		headCommit, err := r.Commit(ref.Target())
		require.NoError(t, err)

		// Create the tag
		sig := object.NewSignature("author", "author@domain.tld")
		tag, err := r.NewTag(&object.TagParams{
			Name:    "v0.0.1-test",
			Target:  headCommit.ToObject(),
			Tagger:  sig,
			Message: "v0.0.1-test",
		})
		require.NoError(t, err)
		// assert the returned object
		assert.Equal(t, "v0.0.1-test", tag.Name())
		assert.Equal(t, ref.Target(), tag.Target())
		assert.Equal(t, "v0.0.1-test", tag.Message())

		// Retrieve the tag
		tagRef, err := r.Tag("v0.0.1-test")
		require.NoError(t, err)

		rawTag, err := r.Object(tagRef.Target())
		require.NoError(t, err)
		fetchedTag, err := rawTag.AsTag()
		require.NoError(t, err)

		assert.Equal(t, tag.ID(), fetchedTag.ID())
		assert.Equal(t, "v0.0.1-test", fetchedTag.Name())
		assert.Equal(t, ref.Target(), fetchedTag.Target())
	})

	t.Run("should fail creating a tag that already exist", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(ginternals.LocalBranchFullName(ginternals.Master))
		require.NoError(t, err)

		headCommit, err := r.Commit(ref.Target())
		require.NoError(t, err)

		// Create the tag
		sig := object.NewSignature("author", "author@domain.tld")
		_, err = r.NewTag(&object.TagParams{
			Name:    "annotated",
			Target:  headCommit.ToObject(),
			Tagger:  sig,
			Message: "annotated",
		})
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrTagExists))
	})

	t.Run("should fail creating a tag using a non-persisted object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		blob := object.New(object.TypeBlob, []byte(""))

		// Create the tag
		sig := object.NewSignature("author", "author@domain.tld")
		_, err = r.NewTag(&object.TagParams{
			Name:    "invalid",
			Target:  blob,
			Tagger:  sig,
			Message: "incvalid",
		})
		require.Error(t, err)
		require.True(t, errors.Is(err, object.ErrObjectInvalid), "got: %s", err.Error())
	})
}

func TestRepositoryNewLightweightTag(t *testing.T) {
	t.Parallel()

	t.Run("create a new valid tag", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(ginternals.LocalBranchFullName(ginternals.Master))
		require.NoError(t, err)

		// Create the tag
		tagRef, err := r.NewLightweightTag("v0.0.1-test", ref.Target())
		require.NoError(t, err)
		// assert the returned object
		assert.Equal(t, ref.Target(), tagRef.Target())
	})

	t.Run("should fail creating a tag that already exist", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(ginternals.LocalBranchFullName(ginternals.Master))
		require.NoError(t, err)

		// Create the tag
		_, err = r.NewLightweightTag("lightweight", ref.Target())
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrTagExists))
	})

	t.Run("should fail creating a tag using a non-persisted object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		blob := object.New(object.TypeBlob, []byte(""))

		// Create the tag
		_, err = r.NewLightweightTag("v0.0.1-test", blob.ID())
		require.Error(t, err)
		require.True(t, errors.Is(err, object.ErrObjectInvalid), "got: %s", err.Error())
	})
}

func TestNewReference(t *testing.T) {
	t.Parallel()

	t.Run("should create a new reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		ref, err := r.NewReference("refs/tests/my_ref", oid)
		require.NoError(t, err)

		assert.Equal(t, ginternals.OidReference, ref.Type())
		assert.Equal(t, "refs/tests/my_ref", ref.Name())
		assert.Equal(t, oid, ref.Target())
		assert.Empty(t, ref.SymbolicTarget())

		ref, err = r.Reference("refs/tests/my_ref")
		require.NoError(t, err)
		assert.Equal(t, ginternals.OidReference, ref.Type())
		assert.Equal(t, "refs/tests/my_ref", ref.Name())
		assert.Equal(t, oid, ref.Target())
		assert.Empty(t, ref.SymbolicTarget())
	})

	t.Run("should overwrite a reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		ref, err := r.NewReference("HEAD", oid)
		require.NoError(t, err)

		assert.Equal(t, ginternals.OidReference, ref.Type())
		assert.Equal(t, "HEAD", ref.Name())
		assert.Equal(t, oid, ref.Target())
		assert.Empty(t, ref.SymbolicTarget())

		ref, err = r.Reference("HEAD")
		require.NoError(t, err)
		assert.Equal(t, ginternals.OidReference, ref.Type())
		assert.Equal(t, "HEAD", ref.Name())
		assert.Equal(t, oid, ref.Target())
		assert.Empty(t, ref.SymbolicTarget())
	})

	t.Run("should fail crating a ref with a name that conflicts", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		_, err = r.NewReference("HEAD/2", oid)
		require.Error(t, err)
	})
}

func TestNewSymbolicReference(t *testing.T) {
	t.Parallel()

	t.Run("should create a new symbolic reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.NewSymbolicReference("refs/tests/my_ref", "refs/heads/master")
		require.NoError(t, err)

		assert.Equal(t, ginternals.SymbolicReference, ref.Type())
		assert.Equal(t, "refs/tests/my_ref", ref.Name())
		assert.Equal(t, ginternals.NullOid, ref.Target())
		assert.Equal(t, "refs/heads/master", ref.SymbolicTarget())

		ref, err = r.Reference("refs/tests/my_ref")
		require.NoError(t, err)
		assert.Equal(t, ginternals.SymbolicReference, ref.Type())
		assert.Equal(t, "refs/tests/my_ref", ref.Name())
		assert.Equal(t, "bbb720a96e4c29b9950a4c577c98470a4d5dd089", ref.Target().String())
		assert.Equal(t, "refs/heads/master", ref.SymbolicTarget())
	})
	t.Run("should overwrite a symbolic reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.NewSymbolicReference("HEAD", "refs/heads/master")
		require.NoError(t, err)

		assert.Equal(t, ginternals.SymbolicReference, ref.Type())
		assert.Equal(t, "HEAD", ref.Name())
		assert.Equal(t, ginternals.NullOid, ref.Target())
		assert.Equal(t, "refs/heads/master", ref.SymbolicTarget())

		ref, err = r.Reference("HEAD")
		require.NoError(t, err)
		assert.Equal(t, ginternals.SymbolicReference, ref.Type())
		assert.Equal(t, "HEAD", ref.Name())
		assert.Equal(t, "bbb720a96e4c29b9950a4c577c98470a4d5dd089", ref.Target().String())
		assert.Equal(t, "refs/heads/master", ref.SymbolicTarget())
	})

	t.Run("should fail crating a ref with a name that conflicts", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		_, err = r.NewSymbolicReference("HEAD/2", "refs/heads/master")
		require.Error(t, err)
	})
}

func TestNewReferenceSafe(t *testing.T) {
	t.Parallel()

	t.Run("should create a new reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		ref, err := r.NewReferenceSafe("refs/tests/my_ref", oid)
		require.NoError(t, err)

		assert.Equal(t, ginternals.OidReference, ref.Type())
		assert.Equal(t, "refs/tests/my_ref", ref.Name())
		assert.Equal(t, oid, ref.Target())
		assert.Empty(t, ref.SymbolicTarget())

		ref, err = r.Reference("refs/tests/my_ref")
		require.NoError(t, err)
		assert.Equal(t, ginternals.OidReference, ref.Type())
		assert.Equal(t, "refs/tests/my_ref", ref.Name())
		assert.Equal(t, oid, ref.Target())
		assert.Empty(t, ref.SymbolicTarget())
	})

	t.Run("should fail overwriting existing reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		_, err = r.NewReferenceSafe("HEAD", oid)
		require.Error(t, err)
	})
}

func TestNewSymbolicReferenceSafe(t *testing.T) {
	t.Parallel()

	t.Run("should create a new symbolic reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.NewSymbolicReferenceSafe("refs/tests/my_ref", "refs/heads/master")
		require.NoError(t, err)

		assert.Equal(t, ginternals.SymbolicReference, ref.Type())
		assert.Equal(t, "refs/tests/my_ref", ref.Name())
		assert.Equal(t, ginternals.NullOid, ref.Target())
		assert.Equal(t, "refs/heads/master", ref.SymbolicTarget())

		ref, err = r.Reference("refs/tests/my_ref")
		require.NoError(t, err)
		assert.Equal(t, ginternals.SymbolicReference, ref.Type())
		assert.Equal(t, "refs/tests/my_ref", ref.Name())
		assert.Equal(t, "bbb720a96e4c29b9950a4c577c98470a4d5dd089", ref.Target().String())
		assert.Equal(t, "refs/heads/master", ref.SymbolicTarget())
	})

	t.Run("should fail overwriting existing symbolic reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		_, err = r.NewSymbolicReferenceSafe("HEAD", "refs/heads/master")
		require.Error(t, err)
	})
}
