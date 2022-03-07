package object_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTree(t *testing.T) {
	t.Parallel()

	t.Run("o.AsTree().ToObject() should return the same object", func(t *testing.T) {
		t.Parallel()

		treeSHA := "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3"
		treeID, err := ginternals.NewOidFromStr(treeSHA)
		require.NoError(t, err)

		testFile := fmt.Sprintf("tree_%s", treeSHA)
		content, err := os.ReadFile(filepath.Join(testutil.TestdataPath(t), testFile))
		require.NoError(t, err)

		o := object.New(object.TypeTree, content)
		tree, err := o.AsTree()
		require.NoError(t, err)

		newO := tree.ToObject()
		require.Equal(t, o.ID(), newO.ID())
		require.Equal(t, treeID, newO.ID())
		require.Equal(t, o.Bytes(), newO.Bytes())
	})

	t.Run("Entries should be immutable", func(t *testing.T) {
		t.Parallel()

		blobSHA := "0343d67ca3d80a531d0d163f0078a81c95c9085a"
		blobID, err := ginternals.NewOidFromStr(blobSHA)
		require.NoError(t, err)

		tree := object.NewTree([]object.TreeEntry{
			{
				Mode: object.ModeFile,
				ID:   blobID,
				Path: "blob",
			},
		})

		tree.Entries()[0].ID[0] = 0xe5
		assert.Equal(t, byte(0x03), tree.Entries()[0].ID[0], "should not update entry ID")

		tree.Entries()[0].Path = "nope"
		assert.Equal(t, "blob", tree.Entries()[0].Path, "should not update entry Path")
	})
}

func TestTreeEntry(t *testing.T) {
	t.Parallel()

	id, err := ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
	require.NoError(t, err)
	id2, err := ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
	require.NoError(t, err)

	tree := object.NewTree([]object.TreeEntry{
		{
			Path: "README.md",
			ID:   id,
			Mode: object.ModeFile,
		},
		{
			Path: "src",
			ID:   id2,
			Mode: object.ModeDirectory,
		},
	})

	testCases := []struct {
		desc        string
		path        string
		expectedOid ginternals.Oid
	}{
		{
			desc:        "existing blob",
			path:        "README.md",
			expectedOid: id,
		},
		{
			desc:        "existing tree",
			path:        "src",
			expectedOid: id2,
		},
		{
			desc:        "invalid path",
			path:        "404",
			expectedOid: ginternals.NullOid,
		},
	}
	for i, tc := range testCases {
		tree := tree
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()
			entry, ok := tree.Entry(tc.path)
			if tc.expectedOid == ginternals.NullOid {
				assert.False(t, ok)
				return
			}
			assert.True(t, ok)
			assert.Equal(t, tc.expectedOid, entry.ID)
		})
	}
}

func TestTreeObjectMode(t *testing.T) {
	t.Parallel()

	t.Run("ObjectType()", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc     string
			mode     object.TreeObjectMode
			expected object.Type
		}{
			{
				desc:     "unknown object should be blob",
				mode:     0o644,
				expected: object.TypeBlob,
			},
			{
				desc:     "ModeFile should be a blob",
				mode:     object.ModeFile,
				expected: object.TypeBlob,
			},
			{
				desc:     "ModeExecutable should be a blob",
				mode:     object.ModeExecutable,
				expected: object.TypeBlob,
			},
			{
				desc:     "ModeSymLink should be a blob",
				mode:     object.ModeSymLink,
				expected: object.TypeBlob,
			},
			{
				desc:     "ModeDirectory should be a tree",
				mode:     object.ModeDirectory,
				expected: object.TypeTree,
			},
			{
				desc:     "ModeGitLink should be a commit",
				mode:     object.ModeGitLink,
				expected: object.TypeCommit,
			},
		}
		for i, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.expected, tc.mode.ObjectType())
			})
		}
	})

	t.Run("IsValid()", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc    string
			mode    object.TreeObjectMode
			isValid bool
		}{
			{
				desc:    "0o644 should not be valid",
				mode:    0o644,
				isValid: false,
			},
			{
				desc:    "ModeFile should be valid",
				mode:    object.ModeFile,
				isValid: true,
			},
			{
				desc:    "0o100755 should be valid",
				mode:    0o100755,
				isValid: true,
			},
		}
		for i, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				out := tc.mode.IsValid()
				assert.Equal(t, tc.isValid, out)
			})
		}
	})
}

func TestNewTreeFromObject(t *testing.T) {
	t.Parallel()

	t.Run("should work on a valid tree", func(t *testing.T) {
		t.Parallel()

		// Find a tree
		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close())
		})

		treeID, err := ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)

		o, err := r.Object(treeID)
		require.NoError(t, err, "failed getting the tree")

		_, err = object.NewTreeFromObject(o)
		require.NoError(t, err)
	})

	t.Run("should fail if the object is not a tree", func(t *testing.T) {
		t.Parallel()

		o := object.New(object.TypeCommit, []byte{})
		_, err := object.NewTreeFromObject(o)
		require.Error(t, err)
		assert.ErrorIs(t, err, object.ErrObjectInvalid)
		assert.Contains(t, err.Error(), "is not a tree")
	})

	t.Run("should work on an empty tree", func(t *testing.T) {
		t.Parallel()

		o := object.New(object.TypeTree, []byte{})
		tree, err := object.NewTreeFromObject(o)
		require.NoError(t, err)
		assert.Len(t, tree.Entries(), 0)
	})

	t.Run("parsing failures", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc               string
			data               string
			expectedErrorMatch string
			expectedError      error
		}{
			{
				desc:               "should fail if the tree has invalid content",
				data:               "mode",
				expectedError:      object.ErrTreeInvalid,
				expectedErrorMatch: "could not retrieve the mode",
			},
			{
				desc:               "should fail if the tree has an invalid mode",
				data:               "mode ",
				expectedError:      object.ErrTreeInvalid,
				expectedErrorMatch: "could not parse mode",
			},
			{
				desc:               "should fail if the tree ends after the mode",
				data:               "644 ",
				expectedError:      object.ErrTreeInvalid,
				expectedErrorMatch: "could not retrieve the path of entry",
			},
			{
				desc:               "should fail if the tree has an invalid ID",
				data:               "644 file.go\x00invalid",
				expectedError:      object.ErrTreeInvalid,
				expectedErrorMatch: "not enough space to retrieve the ID of entry",
			},
		}
		for i, tc := range testCases {
			tc := tc
			i := i
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				o := object.New(object.TypeTree, []byte(tc.data))
				_, err := object.NewTreeFromObject(o)
				require.Error(t, err)
				if tc.expectedError != nil {
					assert.ErrorIs(t, err, tc.expectedError)
				}
				if tc.expectedErrorMatch != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorMatch)
				}
			})
		}
	})
}
