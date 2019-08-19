package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			sig, err := NewSignatureFromBytes([]byte(tc.signature))
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
