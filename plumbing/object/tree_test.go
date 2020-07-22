package object_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTree(t *testing.T) {
	t.Run("o.AsTree().ToObject() should return the same object", func(t *testing.T) {
		t.Parallel()

		treeSHA := "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3"
		treeID, err := plumbing.NewOidFromStr(treeSHA)
		require.NoError(t, err)

		testFile := fmt.Sprintf("tree_%s", treeSHA)
		content, err := ioutil.ReadFile(filepath.Join(testhelper.TestdataPath(t), testFile))
		require.NoError(t, err)

		o := object.NewWithID(treeID, object.TypeTree, content)
		tree, err := o.AsTree()
		require.NoError(t, err)

		newO := tree.ToObject()
		require.Equal(t, o.ID(), newO.ID())
		require.Equal(t, o.Bytes(), newO.Bytes())
	})

	t.Run("Entries should be immutable", func(t *testing.T) {
		t.Parallel()

		treeSHA := "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3"
		treeID, err := plumbing.NewOidFromStr(treeSHA)
		require.NoError(t, err)

		blobSHA := "0343d67ca3d80a531d0d163f0078a81c95c9085a"
		blobID, err := plumbing.NewOidFromStr(blobSHA)
		require.NoError(t, err)

		tree := object.NewTreeWithID(treeID, []object.TreeEntry{
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

func TestTreeObjectMode(t *testing.T) {
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
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			out := tc.mode.IsValid()
			assert.Equal(t, tc.isValid, out)
		})
	}
}
