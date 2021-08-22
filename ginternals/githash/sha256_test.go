package githash_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Nivl/git-go/ginternals/githash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSHA256ConvertFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc          string
		id            string
		expectError   bool
		expectedError error
	}{
		{
			desc:        "valid oid should work",
			id:          "ed7002b439e9ac845f22357d822bac1444730fbdb6016d3ec9432297b9ec9f73",
			expectError: false,
		},
		{
			desc:        "invalid char should fail",
			id:          "ed7002b439e9a c845f22357d822bac1444730fbdb6016d3ec9432297b9ec9f73",
			expectError: true,
		},
		{
			desc:          "invalid size should fail",
			id:            "ed7002b439e9ac845f22357d822bac1444730fbdb6016d3ec9432297b",
			expectError:   true,
			expectedError: githash.ErrInvalidOid,
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			meth := githash.NewSHA256()
			oid, err := meth.ConvertFromString(tc.id)
			if tc.expectError {
				require.Error(t, err)
				assert.True(t, oid.IsZero(), "oid should be Zero")
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError), "invalid error returned: %s", err.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.id, oid.String())
		})
	}
}

func TestSHA256NewOidFromChars(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc          string
		id            []byte
		expectError   bool
		expectedError error
	}{
		{
			desc:        "valid oid should work",
			id:          []byte("ed7002b439e9ac845f22357d822bac1444730fbdb6016d3ec9432297b9ec9f73"),
			expectError: false,
		},
		{
			desc:        "invalid char should fail",
			id:          []byte("ed7002b439e9a c845f22357d822bac1444730fbdb6016d3ec9432297b9ec9f73"),
			expectError: true,
		},
		{
			desc:          "invalid size should fail",
			id:            []byte("ed7002b439e9ac845f22357d822bac1444730fbdb6016d3ec94322"),
			expectError:   true,
			expectedError: githash.ErrInvalidOid,
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			meth := githash.NewSHA256()
			oid, err := meth.ConvertFromChars(tc.id)
			if tc.expectError {
				require.Error(t, err)
				assert.True(t, oid.IsZero(), "oid should be Zero")
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError), "invalid error returned: %s", err.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.id, []byte(oid.String()))
		})
	}
}

func TestSHA256ConvertFromBytes(t *testing.T) {
	t.Parallel()

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
			expectedError: githash.ErrInvalidOid,
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			meth := githash.NewSHA256()
			oid, err := meth.ConvertFromBytes(tc.id)
			if tc.expectError {
				require.Error(t, err)
				assert.True(t, oid.IsZero(), "oid should be Zero")
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError), "invalid error returned: %s", err.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.id, oid.Bytes())
			assert.Equal(t, tc.expectedID, oid.String())
		})
	}
}

func TestSHA256Sum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc       string
		content    []byte
		expectedID []byte
	}{
		{
			desc:       "happy path",
			content:    []byte("123456789"),
			expectedID: []byte("15e2b0d3c33891ebb0f1ef609ec419420c20e320ce94c65fbc8c3312448eb225"),
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			meth := githash.NewSHA256()
			oid := meth.Sum(tc.content)
			assert.Equal(t, tc.expectedID, []byte(oid.String()))
		})
	}
}

func TestSHA256Oid(t *testing.T) {
	t.Parallel()

	t.Run("from string", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc          string
			sha           string
			expectedBytes []byte
			isZero        bool
		}{
			{
				desc:          "valid sha should not be zero",
				sha:           "f7c3bc1d808e04732adf679965ccc34ca7ae3441",
				expectedBytes: []byte{0xf7, 0xc3, 0xbc, 0x1d, 0x80, 0x8e, 0x04, 0x73, 0x2a, 0xdf, 0x67, 0x99, 0x65, 0xcc, 0xc3, 0x4c, 0xa7, 0xae, 0x34, 0x41},
				isZero:        false,
			},
			{
				desc:          "Only 0 should be 0",
				sha:           "0000000000000000000000000000000000000000",
				expectedBytes: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				isZero:        true,
			},
		}
		for i, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				meth := githash.NewSHA256()
				sha, err := meth.ConvertFromString(tc.sha)
				require.NoError(t, err)
				assert.Equal(t, tc.isZero, sha.IsZero())
				assert.Equal(t, tc.sha, sha.String())
				assert.Equal(t, tc.expectedBytes, sha.Bytes())
			})
		}
	})
}
