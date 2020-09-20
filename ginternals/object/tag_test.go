package object_test

import (
	"errors"
	"testing"

	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals"
)

func TestNewTag(t *testing.T) {
	t.Run("NewTag with all data sets", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)
		commitOid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		commit, err := r.GetCommit(commitOid)
		require.NoError(t, err)

		tag, err := object.NewTag(&object.TagParams{
			Target:    commit.ToObject(),
			Message:   "message",
			OptGPGSig: "gpgsig",
			Name:      "v10.5.0",
			Tagger:    object.NewSignature("tagger", "tagger@domain.tld"),
		})
		require.NoError(t, err)
		assert.True(t, tag.ID().IsZero(), "")
		assert.Equal(t, commitOid, tag.Target())
		assert.Equal(t, object.TypeCommit, tag.Type())
		assert.Equal(t, "message", tag.Message())
		assert.Equal(t, "v10.5.0", tag.Name())
		assert.Equal(t, "gpgsig", tag.GPGSig())
		assert.Equal(t, "tagger", tag.Tagger().Name)
	})

	t.Run("non-persisted object should fail", func(t *testing.T) {
		t.Parallel()

		blob := object.New(object.TypeBlob, []byte(""))
		_, err := object.NewTag(&object.TagParams{
			Target:    blob,
			Message:   "message",
			OptGPGSig: "gpgsig",
			Name:      "v10.5.0",
			Tagger:    object.NewSignature("tagger", "tagger@domain.tld"),
		})
		require.Error(t, err)
		require.True(t, errors.Is(err, object.ErrObjectInvalid), "invalid error")
	})
}

func TestTagToObject(t *testing.T) {
	t.Run("ToObject should return the raw object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)
		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)

		// Find a tag
		tagRef, err := r.GetTag("annotated")
		require.NoError(t, err)
		rawTag, err := r.GetObject(tagRef.Target())
		require.NoError(t, err)
		tag, err := rawTag.AsTag()
		require.NoError(t, err)

		// Get the object back
		o := tag.ToObject()
		assert.Equal(t, tag.ID(), o.ID())
	})

	t.Run("happy path on NewTag", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)
		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)
		commitOid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		commit, err := r.GetCommit(commitOid)
		require.NoError(t, err)

		tag, err := object.NewTag(&object.TagParams{
			Target:    commit.ToObject(),
			Message:   "message",
			Name:      "v10.5.0",
			OptGPGSig: "-----BEGIN PGP SIGNATURE-----\n\ndata\n-----END PGP SIGNATURE-----",
			Tagger:    object.NewSignature("tagger", "tagger@domain.tld"),
		})
		require.NoError(t, err)

		o := tag.ToObject()
		tag2, err := o.AsTag()
		require.NoError(t, err)

		assert.Equal(t, tag.Message(), tag2.Message())
		assert.Equal(t, tag.Tagger().Name, tag2.Tagger().Name)
		assert.Equal(t, tag.Name(), tag2.Name())
		assert.Equal(t, tag.GPGSig(), tag2.GPGSig())
		assert.Equal(t, tag.Target(), tag2.Target())
	})
}
