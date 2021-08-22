package object_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/githash"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureString(t *testing.T) {
	t.Parallel()

	sig := object.NewSignature("John Doe", "john@domain.tld")
	// for the sake of the test we gonna cheat a little bit and force
	// the time to be UTC. Otherwise the test would not be consistent
	// on everyone's computer
	now := time.Now().UTC()
	sig.Time = now

	expect := fmt.Sprintf("John Doe <john@domain.tld> %d +0000", now.Unix())
	assert.Equal(t, expect, sig.String())
}

func TestNewSignatureFromBytes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc                 string
		signature            string
		expectsError         bool
		expectsErrorMatch    string
		expectedName         string
		expectedEmail        string
		expectedTimestamp    int64
		expectedTzOffsetMult int
	}{
		{
			desc:                 "valid with a negative offset",
			signature:            "Melvin Laplanche <melvin.wont.reply@gmail.com> 1566115917 -0700",
			expectsError:         false,
			expectedName:         "Melvin Laplanche",
			expectedEmail:        "melvin.wont.reply@gmail.com",
			expectedTimestamp:    int64(1566115917),
			expectedTzOffsetMult: -7,
		},
		{
			desc:                 "valid with a positive offset",
			signature:            "Melvin Laplanche <melvin.wont.reply@gmail.com> 1566005917 +0100",
			expectsError:         false,
			expectedName:         "Melvin Laplanche",
			expectedEmail:        "melvin.wont.reply@gmail.com",
			expectedTimestamp:    int64(1566005917),
			expectedTzOffsetMult: 1,
		},
		{
			desc:                 "valid with a 0 offset",
			signature:            "Melvin Laplanche <melvin.wont.reply@gmail.com> 1566005917 +0000",
			expectsError:         false,
			expectedName:         "Melvin Laplanche",
			expectedEmail:        "melvin.wont.reply@gmail.com",
			expectedTimestamp:    int64(1566005917),
			expectedTzOffsetMult: 0,
		},
		{
			desc:              "invalid offset",
			signature:         "Melvin Laplanche <melvin.wont.reply@gmail.com> 1566005917 nope",
			expectsError:      true,
			expectsErrorMatch: "invalid timezone format",
		},
		{
			desc:                 "valid with a single word name",
			signature:            "Melvin <melvin.wont.reply@gmail.com> 1566005917 -0700",
			expectsError:         false,
			expectedName:         "Melvin",
			expectedEmail:        "melvin.wont.reply@gmail.com",
			expectedTimestamp:    int64(1566005917),
			expectedTzOffsetMult: -7,
		},
		{
			desc:                 "valid with a 4 words name",
			signature:            "Melvin Jacques Marcel Laplanche <melvin.wont.reply@gmail.com> 1566005917 -0700",
			expectsError:         false,
			expectedName:         "Melvin Jacques Marcel Laplanche",
			expectedEmail:        "melvin.wont.reply@gmail.com",
			expectedTimestamp:    int64(1566005917),
			expectedTzOffsetMult: -7,
		},
		{
			desc:                 "valid with specialchar in email",
			signature:            "Melvin Laplanche <melvin.wont.reply+filter@gmail.com> 1566005917 -0700",
			expectsError:         false,
			expectedName:         "Melvin Laplanche",
			expectedEmail:        "melvin.wont.reply+filter@gmail.com",
			expectedTimestamp:    int64(1566005917),
			expectedTzOffsetMult: -7,
		},
		{
			desc:              "invalid timestamp",
			signature:         "Melvin Laplanche <melvin.wont.reply@gmail.com> nope -0700",
			expectsError:      true,
			expectsErrorMatch: "invalid timestamp",
		},
		{
			desc:              "invalid email",
			signature:         "Melvin Laplanche melvin.wont.reply@gmail.com 1566005917 -0700",
			expectsError:      true,
			expectsErrorMatch: "signature stopped after the name",
		},
		{
			desc:              "empty sig",
			signature:         "",
			expectsError:      true,
			expectsErrorMatch: "couldn't retrieve the name",
		},
		{
			desc:              "Name only",
			signature:         "Melvin Laplanche",
			expectsErrorMatch: "signature stopped after the name",
			expectsError:      true,
		},
		{
			desc:              "incomplete sig",
			signature:         "Melvin Laplanche <melvin.wont.reply@gmail.com>",
			expectsError:      true,
			expectsErrorMatch: "signature stopped after the email",
		},
		{
			desc:              "Email not closing",
			signature:         "Melvin Laplanche <melvin.wont.reply@gmail.com",
			expectsError:      true,
			expectsErrorMatch: "couldn't retrieve the email",
		},
		{
			desc:              "email opened but no content",
			signature:         "Melvin Laplanche <",
			expectsErrorMatch: "couldn't retrieve the email",
			expectsError:      true,
		},
		{
			desc:              "Missing timestamp",
			signature:         "Melvin Laplanche <melvin.wont.reply@gmail.com> -0700",
			expectsError:      true,
			expectsErrorMatch: "invalid timestamp",
		},
		{
			desc:              "Missing timestamp",
			signature:         "Melvin Laplanche <melvin.wont.reply@gmail.com>  ",
			expectsError:      true,
			expectsErrorMatch: "signature stopped after the timestamp",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			sig, err := object.NewSignatureFromBytes([]byte(tc.signature))
			if tc.expectsError {
				require.Error(t, err, "NewSignatureFromBytes should have failed")

				if tc.expectsErrorMatch != "" {
					assert.Contains(t, err.Error(), tc.expectsErrorMatch)
				}
				return
			}

			require.NoError(t, err, "NewSignatureFromBytes should have succeed")
			assert.Equal(t, tc.expectedName, sig.Name)
			assert.Equal(t, tc.expectedEmail, sig.Email)
			assert.Equal(t, tc.expectedTimestamp, sig.Time.Unix())
			_, tzOffset := sig.Time.Zone()
			assert.Equal(t, tc.expectedTzOffsetMult*3600, tzOffset)
		})
	}
}

func TestSignatureIsZero(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc   string
		sig    object.Signature
		isZero bool
	}{
		{
			desc:   "empty object should be zero",
			sig:    object.Signature{},
			isZero: true,
		},
		{
			desc: "sign with a name should not be zero",
			sig: object.Signature{
				Name: "tester",
			},
			isZero: false,
		},
		{
			desc: "sign with an email should not be zero",
			sig: object.Signature{
				Email: "tester@domain.tld",
			},
			isZero: false,
		},
		{
			desc: "sign with a time should not be zero",
			sig: object.Signature{
				Time: time.Now(),
			},
			isZero: false,
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.isZero, tc.sig.IsZero())
		})
	}
}

func TestNewCommit(t *testing.T) {
	t.Parallel()

	t.Run("NewCommit with all data sets", func(t *testing.T) {
		t.Parallel()

		// Find a tree
		treeOID, err := ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)

		// Find a commit
		parentID, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		ci := object.NewCommit(treeOID, object.NewSignature("author", "email"), &object.CommitOptions{
			ParentsID: []githash.Oid{parentID},
			Message:   "message",
			GPGSig:    "gpgsig",
			Committer: object.NewSignature("committer", "commiter@domain.tld"),
		})
		assert.Equal(t, treeOID, ci.TreeID())
		assert.Equal(t, "message", ci.Message())
		assert.Equal(t, "gpgsig", ci.GPGSig())
		assert.Equal(t, "committer", ci.Committer().Name)
		assert.Equal(t, "author", ci.Author().Name)
		assert.Equal(t, []githash.Oid{parentID}, ci.ParentIDs())
	})

	t.Run("NewCommit with no committer should use the author", func(t *testing.T) {
		t.Parallel()

		// Find a tree
		treeOID, err := ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)

		ci := object.NewCommit(treeOID, object.NewSignature("author", "email"), &object.CommitOptions{})
		assert.Equal(t, "author", ci.Author().Name)
	})
}

func TestCommitToObject(t *testing.T) {
	t.Parallel()

	t.Run("duplicating a commit should work", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)
		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		// Find a commit
		commitOID, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		commit, err := r.GetCommit(commitOID)
		require.NoError(t, err)

		ci := object.NewCommit(commit.TreeID(), commit.Author(), &object.CommitOptions{
			ParentsID: commit.ParentIDs(),
			Message:   commit.Message(),
			GPGSig:    commit.GPGSig(),
			Committer: commit.Committer(),
		})

		o := ci.ToObject()
		_, err = o.Compress()
		require.NoError(t, err)
		// We're expecting the same ID
		assert.Equal(t, commit.ID(), o.ID())

		ci2, err := o.AsCommit()
		require.NoError(t, err)

		// We're expecting the same content
		assert.Equal(t, commit.Message(), ci2.Message())
		assert.Equal(t, commit.Committer().Name, ci2.Committer().Name)
		assert.Equal(t, commit.ParentIDs(), ci2.ParentIDs())
		assert.Equal(t, commit.GPGSig(), ci2.GPGSig())
		assert.Equal(t, commit.TreeID(), ci2.TreeID())
	})

	t.Run("ToObject should return the raw object", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)
		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, r.Close(), "failed closing repo")
		})

		// Find a commit
		commitOID, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		commit, err := r.GetCommit(commitOID)
		require.NoError(t, err)

		o := commit.ToObject()
		assert.Equal(t, commit.ID(), o.ID())

		ci, err := o.AsCommit()
		require.NoError(t, err)

		assert.Equal(t, commit.Message(), ci.Message())
		assert.Equal(t, commit.Committer().Name, ci.Committer().Name)
		assert.Equal(t, commit.ParentIDs(), ci.ParentIDs())
		assert.Equal(t, commit.GPGSig(), ci.GPGSig())
		assert.Equal(t, commit.TreeID(), ci.TreeID())
	})

	t.Run("happy path on NewCommit", func(t *testing.T) {
		t.Parallel()

		// Find a tree
		treeOID, err := ginternals.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)

		// Find a commit
		parentID, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		ci := object.NewCommit(treeOID, object.NewSignature("author", "email"), &object.CommitOptions{
			ParentsID: []githash.Oid{parentID},
			Message:   "message",
			GPGSig:    "-----BEGIN PGP SIGNATURE-----\n\ndata\n-----END PGP SIGNATURE-----",
			Committer: object.NewSignature("committer", "commiter@domain.tld"),
		})

		o := ci.ToObject()
		ci2, err := o.AsCommit()
		require.NoError(t, err)

		assert.Equal(t, ci.Message(), ci2.Message())
		assert.Equal(t, ci.Committer().Name, ci2.Committer().Name)
		assert.Equal(t, ci.ParentIDs(), ci2.ParentIDs())
		assert.Equal(t, ci.GPGSig(), ci2.GPGSig())
		assert.Equal(t, ci.TreeID(), ci2.TreeID())
	})
}

func TestNewCommitFromObject(t *testing.T) {
	t.Parallel()

	t.Run("should work on a valid commit", func(t *testing.T) {
		t.Parallel()

		// Find a commit
		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		t.Cleanup(cleanup)

		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err, "failed loading a repo")
		require.NotNil(t, r, "repository should not be nil")
		t.Cleanup(func() {
			require.NoError(t, r.Close())
		})

		commitID, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		o, err := r.GetObject(commitID)
		require.NoError(t, err, "failed loading a repo")

		_, err = object.NewCommitFromObject(o)
		require.NoError(t, err)
	})

	t.Run("should fail if the object is not a commit", func(t *testing.T) {
		t.Parallel()

		o := object.New(object.TypeBlob, []byte{})
		_, err := object.NewCommitFromObject(o)
		require.Error(t, err)
		assert.ErrorIs(t, err, object.ErrObjectInvalid)
		assert.Contains(t, err.Error(), "is not a commit")
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
				desc:          "should fail if the commit has invalid content",
				data:          "invalid data",
				expectedError: object.ErrCommitInvalid,
			},
			{
				desc:          "should fail if the commit has incomplete content",
				data:          "invalid data\n",
				expectedError: object.ErrCommitInvalid,
			},
			{
				desc:               "should fail if the tree id is invalid",
				data:               "tree adad\n",
				expectedErrorMatch: "could not parse tree id",
			},
			{
				desc:               "should fail if the parent id is invalid",
				data:               "parent adad\n",
				expectedErrorMatch: "could not parse parent id",
			},
			{
				desc:               "should fail if the author is invalid",
				data:               "author adad\n",
				expectedErrorMatch: "could not parse author signature",
			},
			{
				desc:               "should fail if the committer is invalid",
				data:               "committer adad\n",
				expectedErrorMatch: "could not parse committer signature",
			},
		}
		for i, tc := range testCases {
			tc := tc
			i := i
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				o := object.New(object.TypeCommit, []byte(tc.data))
				_, err := object.NewCommitFromObject(o)
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
