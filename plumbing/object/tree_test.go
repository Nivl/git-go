package object_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/stretchr/testify/require"
)

func TestTreeToObject(t *testing.T) {
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

		newO, err := tree.ToObject()
		require.NoError(t, err)
		require.Equal(t, o.ID(), newO.ID())
		require.Equal(t, o.Bytes(), newO.Bytes())
	})
}
