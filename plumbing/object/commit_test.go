package object_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureString(t *testing.T) {
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
	testCases := []struct {
		desc                 string
		signature            string
		expectsError         bool
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
			desc:         "invalid offset",
			signature:    "Melvin Laplanche <melvin.wont.reply@gmail.com> 1566005917 nope",
			expectsError: true,
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
			desc:         "invalid timestamp",
			signature:    "Melvin Laplanche <melvin.wont.reply@gmail.com> nope -0700",
			expectsError: true,
		},
		{
			desc:         "invalid email",
			signature:    "Melvin Laplanche melvin.wont.reply@gmail.com 1566005917 -0700",
			expectsError: true,
		},
		{
			desc:         "empty sig",
			signature:    "",
			expectsError: true,
		},
		{
			desc:         "incomplete sig",
			signature:    "Melvin Laplanche <melvin.wont.reply@gmail.com>",
			expectsError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			sig, err := object.NewSignatureFromBytes([]byte(tc.signature))
			if tc.expectsError {
				require.Error(t, err, "NewSignatureFromBytes should have failed")
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
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.isZero, tc.sig.IsZero())
		})
	}
}

func TestNewCommit(t *testing.T) {
	t.Run("NewCommit with all data sets", func(t *testing.T) {
		t.Parallel()

		// Find a tree
		treeOID, err := plumbing.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)

		// Find a commit
		parentID, err := plumbing.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		ci := object.NewCommit(treeOID, object.NewSignature("author", "email"), &object.CommitOptions{
			ParentsID: []plumbing.Oid{parentID},
			Message:   "message",
			GPGSig:    "gpgsig",
			Committer: object.NewSignature("committer", "commiter@domain.tld"),
		})
		assert.Equal(t, treeOID, ci.TreeID())
		assert.Equal(t, "message", ci.Message())
		assert.Equal(t, "gpgsig", ci.GPGSig())
		assert.Equal(t, "committer", ci.Committer().Name)
		assert.Equal(t, "author", ci.Author().Name)
		assert.Equal(t, []plumbing.Oid{parentID}, ci.ParentIDs())
	})

	t.Run("NewCommit with no committer should use the author", func(t *testing.T) {
		t.Parallel()

		// Find a tree
		treeOID, err := plumbing.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)

		ci := object.NewCommit(treeOID, object.NewSignature("author", "email"), &object.CommitOptions{})
		assert.Equal(t, "author", ci.Author().Name)
	})
}

func TestToObject(t *testing.T) {
	t.Run("duplicating a commit should work", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
		defer cleanup()
		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)

		// Find a commit
		commitOID, err := plumbing.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
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
		defer cleanup()
		r, err := git.OpenRepository(repoPath)
		require.NoError(t, err)

		// Find a commit
		commitOID, err := plumbing.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
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
		treeOID, err := plumbing.NewOidFromStr("e5b9e846e1b468bc9597ff95d71dfacda8bd54e3")
		require.NoError(t, err)

		// Find a commit
		parentID, err := plumbing.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)

		ci := object.NewCommit(treeOID, object.NewSignature("author", "email"), &object.CommitOptions{
			ParentsID: []plumbing.Oid{parentID},
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
