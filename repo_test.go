package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/gitpath"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Parallel()

	t.Run("repo with working tree", func(t *testing.T) {
		t.Parallel()

		// Setup
		d, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		// Run logic
		r, err := InitRepository(d)
		require.NoError(t, err, "failed creating a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		// assert returned repository
		assert.Equal(t, d, r.repoRoot)
		assert.Equal(t, filepath.Join(d, gitpath.DotGitPath), r.dotGitPath)
		assert.NotNil(t, r.wt)
		assert.False(t, r.IsBare(), "repos should not be bare")
	})

	t.Run("bare repo", func(t *testing.T) {
		t.Parallel()

		// Setup
		d, cleanup := testhelper.TempDir(t)
		t.Cleanup(cleanup)

		// Run logic
		r, err := InitRepositoryWithOptions(d, InitOptions{
			IsBare: true,
		})
		require.NoError(t, err, "failed creating a repo")

		// assert returned repository
		require.Equal(t, d, r.repoRoot)
		require.Equal(t, d, r.dotGitPath)
		assert.Nil(t, r.wt)
		assert.True(t, r.IsBare(), "repos should be bare")
	})
}

func TestOpen(t *testing.T) {
	t.Parallel()

	t.Run("repo with working tree", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close())
		})

		// assert returned repository
		assert.Equal(t, repoPath, r.repoRoot)
		assert.Equal(t, filepath.Join(repoPath, gitpath.DotGitPath), r.dotGitPath)
		assert.NotNil(t, r.wt)
		assert.False(t, r.IsBare(), "repos should not be bare")
	})

	t.Run("bare repo", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)
		repoPath = filepath.Join(repoPath, gitpath.DotGitPath)

		r, err := OpenRepositoryWithOptions(repoPath, OpenOptions{
			IsBare: true,
		})
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		// assert returned repository
		// assert returned repository
		require.Equal(t, repoPath, r.repoRoot)
		require.Equal(t, repoPath, r.dotGitPath)
		assert.Nil(t, r.wt)
		assert.True(t, r.IsBare(), "repos should be bare")
	})
}

func TestRepositoryGetObject(t *testing.T) {
	t.Parallel()

	t.Run("loose object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("b07e28976ac8972715598f390964d53cf4dbc1bd")
		require.NoError(t, err)

		obj, err := r.GetObject(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID())
		assert.Equal(t, object.TypeBlob, obj.Type())
		assert.Equal(t, "package packfile", string(obj.Bytes()[:16]))
	})

	t.Run("Object from packfile", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("1dcdadc2a420225783794fbffd51e2e137a69646")
		require.NoError(t, err)

		obj, err := r.GetObject(oid)
		require.NoError(t, err)
		require.NotNil(t, obj)

		assert.Equal(t, oid, obj.ID())
		assert.Equal(t, object.TypeCommit, obj.Type())
	})
}

func TestRepositoryNewBlob(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
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
	p := filepath.Join(r.dotGitPath, gitpath.ObjectsPath, blob.ID().String()[0:2], blob.ID().String()[2:])
	_, err = os.Stat(p)
	require.NoError(t, err)
}

func TestRepositoryGetCommit(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)

	r, err := OpenRepository(repoPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	commitOid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
	require.NoError(t, err)

	c, err := r.GetCommit(commitOid)
	require.NoError(t, err)

	assert.Equal(t, commitOid, c.ID())
	assert.Equal(t, "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3", c.TreeID().String())
	require.Len(t, c.ParentIDs(), 1)
	assert.Equal(t, "6097a04b7a327c4be68f222ca66e61b8e1abe5c1", c.ParentIDs()[0].String())
}

func TestRepositoryGetReference(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
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

			ref, err := r.GetReference(tc.refName)

			if tc.expectedError != nil {
				assert.True(t, errors.Is(err, tc.expectedError), "wrong error returned")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedTarget, ref.Target().String())
		})
	}
}

func TestRepositoryGetTree(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)

	r, err := OpenRepository(repoPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	treeOid, err := ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
	require.NoError(t, err)

	tree, err := r.GetTree(treeOid)
	require.NoError(t, err)

	assert.Equal(t, treeOid, tree.ID())
	require.Len(t, tree.Entries(), 13)
}

func TestRepositoryNewCommit(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)

	r, err := OpenRepository(repoPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	ref, err := r.dotGit.Reference(gitpath.LocalBranch(ginternals.Master))
	require.NoError(t, err)

	headCommit, err := r.GetCommit(ref.Target())
	require.NoError(t, err)

	headTree, err := r.GetTree(headCommit.TreeID())
	require.NoError(t, err)

	sig := object.NewSignature("author", "author@domain.tld")
	c, err := r.NewCommit(gitpath.LocalBranch(ginternals.Master), headTree, sig, &object.CommitOptions{
		ParentsID: []ginternals.Oid{headCommit.ID()},
		Message:   "new commit that doesn't do anything",
	})
	require.NoError(t, err)

	// The commit should be findable
	_, err = r.GetCommit(c.ID())
	require.NoError(t, err)

	// We update the ref since it should have changed
	ref, err = r.dotGit.Reference(gitpath.LocalBranch(ginternals.Master))
	require.NoError(t, err)
	assert.Equal(t, c.ID(), ref.Target())
}

func TestRepositoryNewDetachedCommit(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)

	r, err := OpenRepository(repoPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	ref, err := r.dotGit.Reference(gitpath.LocalBranch(ginternals.Master))
	require.NoError(t, err)

	headCommit, err := r.GetCommit(ref.Target())
	require.NoError(t, err)

	headTree, err := r.GetTree(headCommit.TreeID())
	require.NoError(t, err)

	sig := object.NewSignature("author", "author@domain.tld")
	c, err := r.NewDetachedCommit(headTree, sig, &object.CommitOptions{
		ParentsID: []ginternals.Oid{headCommit.ID()},
		Message:   "new commit that doesn't do anything",
	})
	require.NoError(t, err)

	// The commit should be findable
	_, err = r.GetCommit(c.ID())
	require.NoError(t, err)

	// We update the ref to make sure it's not updated
	updateddRef, err := r.dotGit.Reference(gitpath.LocalBranch(ginternals.Master))
	require.NoError(t, err)
	assert.Equal(t, ref.Target(), updateddRef.Target())
}

func TestRepositoryGetTag(t *testing.T) {
	t.Parallel()

	t.Run("annotated", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		tagID, err := ginternals.NewOidFromStr("80316e01dbfdf5c2a8a20de66c747ecd4c4bd442")
		require.NoError(t, err)

		tagRef, err := r.GetTag("annotated")
		require.NoError(t, err)

		require.Equal(t, tagID, tagRef.Target())

		rawTag, err := r.GetObject(tagRef.Target())
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

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		targettedCommitID, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		tagRef, err := r.GetTag("lightweight")
		require.NoError(t, err)

		require.Equal(t, targettedCommitID, tagRef.Target())

		commit, err := r.GetCommit(tagRef.Target())
		require.NoError(t, err)

		assert.Equal(t, targettedCommitID, commit.ID())
	})

	t.Run("unexisting tag", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		_, err = r.GetTag("does-not-exist")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrTagNotFound), "invalid error type")
	})
}

func TestRepositoryNewTag(t *testing.T) {
	t.Parallel()

	t.Run("create a new valid tag", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(gitpath.LocalBranch(ginternals.Master))
		require.NoError(t, err)

		headCommit, err := r.GetCommit(ref.Target())
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
		tagRef, err := r.GetTag("v0.0.1-test")
		require.NoError(t, err)

		rawTag, err := r.GetObject(tagRef.Target())
		require.NoError(t, err)
		fetchedTag, err := rawTag.AsTag()
		require.NoError(t, err)

		assert.Equal(t, tag.ID(), fetchedTag.ID())
		assert.Equal(t, "v0.0.1-test", fetchedTag.Name())
		assert.Equal(t, ref.Target(), fetchedTag.Target())
	})

	t.Run("should fail creating a tag that already exist", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(gitpath.LocalBranch(ginternals.Master))
		require.NoError(t, err)

		headCommit, err := r.GetCommit(ref.Target())
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

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
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
		require.True(t, errors.Is(err, object.ErrObjectInvalid))
	})
}

func TestRepositoryNewLightweightTag(t *testing.T) {
	t.Parallel()

	t.Run("create a new valid tag", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(gitpath.LocalBranch(ginternals.Master))
		require.NoError(t, err)

		// Create the tag
		tagRef, err := r.NewLightweightTag("v0.0.1-test", ref.Target())
		require.NoError(t, err)
		// assert the returned object
		assert.Equal(t, ref.Target(), tagRef.Target())
	})

	t.Run("should fail creating a tag that already exist", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		ref, err := r.dotGit.Reference(gitpath.LocalBranch(ginternals.Master))
		require.NoError(t, err)

		// Create the tag
		_, err = r.NewLightweightTag("lightweight", ref.Target())
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrTagExists))
	})

	t.Run("should fail creating a tag using a non-persisted object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
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
		require.True(t, errors.Is(err, object.ErrObjectInvalid))
	})
}
