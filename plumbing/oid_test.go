package plumbing_test

import (
	"fmt"
	"testing"

	"github.com/Nivl/git-go/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestNewOidFromStr(t *testing.T) {
	testCases := []struct {
		desc          string
		id            string
		expectError   bool
		expectedError error
	}{
		{
			desc:        "valid oid should work",
			id:          "0eaf966ff79d8f61958aaefe163620d952606516",
			expectError: false,
		},
		{
			desc:        "invalid char should fail",
			id:          "0eaf96 ff79d8f61958aaefe163620d952606516",
			expectError: true,
		},
		{
			desc:          "invalid size should fail",
			id:            "0eaf96ff79d8f61958aaefe163620d952606",
			expectError:   true,
			expectedError: plumbing.ErrInvalidOid,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			oid, err := plumbing.NewOidFromStr(tc.id)
			if tc.expectError {
				require.Error(t, err)
				assert.Equal(t, plumbing.NullOid, oid)
				if tc.expectedError != nil {
					assert.True(t, xerrors.Is(err, plumbing.ErrInvalidOid), "invalid error returned: %s", err.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.id, oid.String())
		})
	}
}

func TestNewOidFromChars(t *testing.T) {
	testCases := []struct {
		desc          string
		id            []byte
		expectError   bool
		expectedError error
	}{
		{
			desc:        "valid oid should work",
			id:          []byte("0eaf966ff79d8f61958aaefe163620d952606516"),
			expectError: false,
		},
		{
			desc:        "invalid char should fail",
			id:          []byte("0eaf96 ff79d8f61958aaefe163620d952606516"),
			expectError: true,
		},
		{
			desc:          "invalid size should fail",
			id:            []byte("0eaf96ff79d8f61958aaefe163620d952606"),
			expectError:   true,
			expectedError: plumbing.ErrInvalidOid,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			oid, err := plumbing.NewOidFromChars(tc.id)
			if tc.expectError {
				require.Error(t, err)
				assert.Equal(t, plumbing.NullOid, oid)
				if tc.expectedError != nil {
					assert.True(t, xerrors.Is(err, plumbing.ErrInvalidOid), "invalid error returned: %s", err.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.id, []byte(oid.String()))
		})
	}
}

func TestNewOidFromHex(t *testing.T) {
	testCases := []struct {
		desc          string
		id            []byte
		expectedID    string
		expectError   bool
		expectedError error
	}{
		{
			desc:        "valid oid should work",
			id:          []byte{0x0e, 0xaf, 0x96, 0x6f, 0xf7, 0x9d, 0x8f, 0x61, 0x95, 0x8a, 0xae, 0xfe, 0x16, 0x36, 0x20, 0xd9, 0x52, 0x60, 0x65, 0x16},
			expectError: false,
			expectedID:  "0eaf966ff79d8f61958aaefe163620d952606516",
		},
		{
			desc:          "invalid size should fail",
			id:            []byte{0x0e, 0xaf, 0x96, 0x6f, 0xf7, 0x9d, 0x8f, 0x61, 0x95, 0x8a, 0xae, 0xfe, 0x16, 0x36, 0x20, 0xd9, 0x52, 0x60, 0x65},
			expectError:   true,
			expectedError: plumbing.ErrInvalidOid,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			oid, err := plumbing.NewOidFromHex(tc.id)
			if tc.expectError {
				require.Error(t, err)
				assert.Equal(t, plumbing.NullOid, oid)
				if tc.expectedError != nil {
					assert.True(t, xerrors.Is(err, plumbing.ErrInvalidOid), "invalid error returned: %s", err.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.id, oid.Bytes())
			assert.Equal(t, tc.expectedID, oid.String())
		})
	}
}

func TestNewOidFromContent(t *testing.T) {
	testCases := []struct {
		desc       string
		content    []byte
		expectedID []byte
	}{
		{
			desc:       "happy path",
			content:    []byte("123456789"),
			expectedID: []byte("f7c3bc1d808e04732adf679965ccc34ca7ae3441"),
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			oid := plumbing.NewOidFromContent(tc.content)
			assert.Equal(t, tc.expectedID, []byte(oid.String()))
		})
	}
}

func TestIsZero(t *testing.T) {
	t.Run("from string", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc   string
			sha    string
			isZero bool
		}{
			{
				desc:   "valid sha should not be zero",
				sha:    "f7c3bc1d808e04732adf679965ccc34ca7ae3441",
				isZero: false,
			},
			{
				desc:   "Only 0 should be 0",
				sha:    "0000000000000000000000000000000000000000",
				isZero: true,
			},
		}
		for i, tc := range testCases {
			tc := tc
			i := i
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				sha, err := plumbing.NewOidFromStr(tc.sha)
				require.NoError(t, err)
				require.Equal(t, tc.isZero, sha.IsZero())
			})
		}
	})

	t.Run("NullOid should be nul", func(t *testing.T) {
		t.Parallel()
		require.True(t, plumbing.NullOid.IsZero(), "NullOid should be Zero")
	})
}
