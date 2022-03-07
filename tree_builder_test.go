package git

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeBuilderInsert(t *testing.T) {
	t.Parallel()

	t.Run("single pass/fail", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc          string
			sha           string
			expectedError error
		}{
			{
				desc:          "should fail inserting an object that doesn't exist",
				sha:           ginternals.NullOid.String(),
				expectedError: ginternals.ErrObjectNotFound,
			},
			{
				desc:          "should fail inserting a commit",
				sha:           "bbb720a96e4c29b9950a4c577c98470a4d5dd089",
				expectedError: object.ErrObjectInvalid,
			},
			{
				desc: "should pass inserting a blob",
				sha:  "642480605b8b0fd464ab5762e044269cf29a60a3",
			},
			{
				desc: "should pass inserting a tree",
				sha:  "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3",
			},
		}
		for i, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
				t.Cleanup(cleanup)

				r, err := OpenRepository(repoPath)
				require.NoError(t, err, "failed loading a repo")
				require.NotNil(t, r, "repository should not be nil")
				t.Cleanup(func() {
					require.NoError(t, r.Close(), "failed closing repo")
				})

				oid, err := ginternals.NewOidFromStr(tc.sha)
				require.NoError(t, err)

				tb := r.NewTreeBuilder()
				err = tb.Insert("somewhere", oid, object.ModeFile)
				if tc.expectedError != nil {
					require.Error(t, err)
					assert.True(t, errors.Is(err, tc.expectedError))
				} else {
					require.NoError(t, err)
					assert.Len(t, tb.entries, 1)
				}
			})
		}
	})

	t.Run("should pass inserting multiple objects", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		tb := r.NewTreeBuilder()

		// insert a blob
		oid, err := ginternals.NewOidFromStr("642480605b8b0fd464ab5762e044269cf29a60a3")
		require.NoError(t, err)
		err = tb.Insert("blob", oid, object.ModeFile)
		require.NoError(t, err)

		// insert a tree
		oid, err = ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)
		err = tb.Insert("tree", oid, object.ModeDirectory)
		require.NoError(t, err)

		assert.Len(t, tb.entries, 2)
	})

	t.Run("should pass overwritting a path", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		tb := r.NewTreeBuilder()

		// insert a blob
		oid, err := ginternals.NewOidFromStr("642480605b8b0fd464ab5762e044269cf29a60a3")
		require.NoError(t, err)
		err = tb.Insert("path", oid, object.ModeFile)
		require.NoError(t, err)

		// insert a tree
		oid, err = ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)
		err = tb.Insert("path", oid, object.ModeDirectory)
		require.NoError(t, err)

		assert.Len(t, tb.entries, 1)
		require.Contains(t, tb.entries, "path")
		require.Equal(t, tb.entries["path"].ID, oid)
		require.Equal(t, tb.entries["path"].Mode, object.ModeDirectory)
	})

	t.Run("should fail with invalid mode", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		tb := r.NewTreeBuilder()

		// insert a blob
		oid, err := ginternals.NewOidFromStr("642480605b8b0fd464ab5762e044269cf29a60a3")
		require.NoError(t, err)
		err = tb.Insert("path", oid, 0o644)
		require.Error(t, err)
	})
}

func TestTreeBuilderRemove(t *testing.T) {
	t.Parallel()

	t.Run("should remove elements", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		tb := r.NewTreeBuilder()

		// insert a blob
		oid, err := ginternals.NewOidFromStr("642480605b8b0fd464ab5762e044269cf29a60a3")
		require.NoError(t, err)
		err = tb.Insert("blob", oid, object.ModeFile)
		require.NoError(t, err)

		// insert a tree
		oid, err = ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)
		err = tb.Insert("tree", oid, object.ModeDirectory)
		require.NoError(t, err)
		assert.Len(t, tb.entries, 2)

		// Remove the blob
		tb.Remove("blob")
		assert.Len(t, tb.entries, 1)

		// Remove the tree
		tb.Remove("tree")
		assert.Len(t, tb.entries, 0)
	})

	t.Run("should pass removing something that doesn't exists", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		tb := r.NewTreeBuilder()

		// Remove the blob
		assert.Len(t, tb.entries, 0)
		tb.Remove("blob")
		assert.Len(t, tb.entries, 0)

		// Let's test with an allocated map
		tb.entries = map[string]object.TreeEntry{}
		tb.Remove("blob")
		assert.Len(t, tb.entries, 0)
	})
}

func TestTreeBuilderWrite(t *testing.T) {
	t.Parallel()

	t.Run("should return 4b825dc642cb6eb9a060e54bf8d69288fbee4904 for empty tree", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		tb := r.NewTreeBuilder()
		tree, err := tb.Write()
		require.NoError(t, err)
		assert.Empty(t, tree.Entries())
		assert.Equal(t, "4b825dc642cb6eb9a060e54bf8d69288fbee4904", tree.ID().String())
	})

	t.Run("should persist tree", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		tb := r.NewTreeBuilder()

		// insert a blob
		oid, err := ginternals.NewOidFromStr("642480605b8b0fd464ab5762e044269cf29a60a3")
		require.NoError(t, err)
		err = tb.Insert("blob", oid, object.ModeFile)
		require.NoError(t, err)

		// insert a tree
		oid, err = ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)
		err = tb.Insert("tree", oid, object.ModeDirectory)
		require.NoError(t, err)

		tree, err := tb.Write()
		require.NoError(t, err)
		assert.Len(t, tb.entries, 2)

		p := ginternals.LooseObjectPath(r.Config, tree.ID().String())
		assert.FileExists(t, p)
	})

	t.Run("building an existing tree should return the same data", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		oid, err := ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)
		o, err := r.Object(oid)
		require.NoError(t, err)
		tree, err := o.AsTree()
		require.NoError(t, err)

		// Create a tree and write it right away
		tb := r.NewTreeBuilderFromTree(tree)
		newTree, err := tb.Write()
		require.NoError(t, err)
		assert.Equal(t, tree.ID().String(), newTree.ID().String())
		assert.Equal(t, tree.Entries(), newTree.Entries())
	})
}
