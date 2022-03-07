package object_test

import (
	"fmt"
	"testing"

	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals"
)

func TestNewTag(t *testing.T) {
	t.Parallel()

	t.Run("NewTag with all data sets", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})
		commitOid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		commit, err := r.Commit(commitOid)
		require.NoError(t, err)

		tag := object.NewTag(&object.TagParams{
			Target:    commit.ToObject(),
			Message:   "message",
			OptGPGSig: "gpgsig",
			Name:      "v10.5.0",
			Tagger:    object.NewSignature("tagger", "tagger@domain.tld"),
		})
		assert.True(t, tag.ID().IsZero(), "")
		assert.Equal(t, commitOid, tag.Target())
		assert.Equal(t, object.TypeCommit, tag.Type())
		assert.Equal(t, "message", tag.Message())
		assert.Equal(t, "v10.5.0", tag.Name())
		assert.Equal(t, "gpgsig", tag.GPGSig())
		assert.Equal(t, "tagger", tag.Tagger().Name)
	})
}

func TestTagToObject(t *testing.T) {
	t.Parallel()

	t.Run("ToObject should return the raw object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)
		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		// Find a tag
		tagRef, err := r.Tag("annotated")
		require.NoError(t, err)
		rawTag, err := r.Object(tagRef.Target())
		require.NoError(t, err)
		tag, err := rawTag.AsTag()
		require.NoError(t, err)

		// Get the object back
		o := tag.ToObject()
		assert.Equal(t, tag.ID(), o.ID())
	})

	t.Run("happy path on NewTag", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)
		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})
		commitOid, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		commit, err := r.Commit(commitOid)
		require.NoError(t, err)

		tag := object.NewTag(&object.TagParams{
			Target:    commit.ToObject(),
			Message:   "message",
			Name:      "v10.5.0",
			OptGPGSig: "-----BEGIN PGP SIGNATURE-----\n\ndata\n-----END PGP SIGNATURE-----",
			Tagger:    object.NewSignature("tagger", "tagger@domain.tld"),
		})

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

func TestNewTagFromObject(t *testing.T) {
	t.Parallel()

	t.Run("should work on a valid tag", func(t *testing.T) {
		t.Parallel()

		// Find a tag
		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close())
		})

		tagRef, err := r.Tag("annotated")
		require.NoError(t, err)

		o, err := r.Object(tagRef.Target())
		require.NoError(t, err, "failed fetching a tag")

		_, err = object.NewTagFromObject(o)
		require.NoError(t, err)
	})

	t.Run("should fail if the object is not a tag", func(t *testing.T) {
		t.Parallel()

		o := object.New(object.TypeTree, []byte{})
		_, err := object.NewTagFromObject(o)
		require.Error(t, err)
		assert.ErrorIs(t, err, object.ErrObjectInvalid)
		assert.Contains(t, err.Error(), "is not a tag")
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
				desc:          "should fail if the tag has invalid content",
				data:          "invalid data",
				expectedError: object.ErrTagInvalid,
			},
			{
				desc:               "should fail if the tag has incomplete content",
				data:               "invalid data\n",
				expectedError:      object.ErrTagInvalid,
				expectedErrorMatch: "tag has no tagger",
			},
			{
				desc:               "should fail if the object id is invalid",
				data:               "object adad\n",
				expectedErrorMatch: "could not parse target id",
			},
			{
				desc:               "should fail if the object id is invalid",
				data:               "type nope\n",
				expectedErrorMatch: "invalid object type",
			},
			{
				desc:               "should fail if the object id is invalid",
				data:               "tagger nope\n",
				expectedErrorMatch: "could not parse tagger",
			},
		}
		for i, tc := range testCases {
			tc := tc
			i := i
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				o := object.New(object.TypeTag, []byte(tc.data))
				_, err := object.NewTagFromObject(o)
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
