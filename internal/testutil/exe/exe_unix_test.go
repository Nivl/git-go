//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package exe_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Nivl/git-go/internal/testutil/exe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		cmd            string
		args           []string
		expectedOutput string
		expectedError  error
	}{
		{
			cmd:            "echo",
			args:           []string{"this", "should be printed"},
			expectedOutput: "this should be printed",
			expectedError:  nil,
		},
		{
			cmd:            "does-not-exist",
			args:           []string{},
			expectedOutput: "",
			expectedError:  errors.New(`exec: "does-not-exist": executable file not found in $PATH`),
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s %s", i, tc.cmd, tc.args), func(t *testing.T) {
			t.Parallel()

			out, err := exe.Run(tc.cmd, tc.args...)
			if tc.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, out)
		})
	}
}
