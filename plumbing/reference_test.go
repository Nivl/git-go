package plumbing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRefNameValid(t *testing.T) {
	testCases := []struct {
		desc       string
		name       string
		shouldPass bool
	}{
		{
			desc:       "name with control chars should fail",
			name:       "ml/not\000valide",
			shouldPass: false,
		},
		{
			desc:       "name with control chars should fail",
			name:       "ml/not\177valide",
			shouldPass: false,
		},
		{
			desc:       "name with slashes should pass",
			name:       "ml/some/name_/that/I/often-use/89",
			shouldPass: true,
		},
		{
			desc:       "name cannot be empty",
			name:       "",
			shouldPass: false,
		},
		{
			desc:       "name cannot start with a /",
			name:       "/refs/heads/master",
			shouldPass: false,
		},
		{
			desc:       "name cannot end with a /",
			name:       "refs/heads/master/",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain ..",
			name:       "refs/heads/ma..ster",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain ?",
			name:       "refs/heads/master?",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain :",
			name:       "refs/heads/ma:ster",
			shouldPass: false,
		},
		{
			desc:       `name cannot contain \`,
			name:       `refs/heads/ma\ster`,
			shouldPass: false,
		},
		{
			desc:       "name cannot contain ^",
			name:       "refs/heads/ma^ster",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain @{",
			name:       "refs/heads/ma@{ster}",
			shouldPass: false,
		},
		{
			desc:       "name can end with @",
			name:       "refs/heads/master@",
			shouldPass: true,
		},
		{
			desc:       "name cannot start with a .",
			name:       ".refs/heads/master",
			shouldPass: false,
		},
		{
			desc:       "name cannot end with a .",
			name:       "refs/heads/master.",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain a [",
			name:       "[refs/heads/master",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain a space",
			name:       "refs/he ads/master",
			shouldPass: false,
		},
		{
			desc:       "name cannot end with .lock",
			name:       "refs/heads/master.lock",
			shouldPass: false,
		},
		{
			desc:       "segments cannot be empty",
			name:       "refs//master",
			shouldPass: false,
		},
		{
			desc:       "segments cannot end with a .",
			name:       "refs/heads./master",
			shouldPass: false,
		},
		{
			desc:       "segments cannot end with .lock",
			name:       "refs/heads.lock/master",
			shouldPass: false,
		},
		{
			desc:       "HEAD should be a valid reference",
			name:       "HEAD",
			shouldPass: true,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			res := IsRefNameValid(tc.name)
			assert.Equal(t, tc.shouldPass, res)
		})
	}
}
