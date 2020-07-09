// Package testhelper contains helpers to simplify tests
package testhelper

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TempDir creates a temp dir and returns a cleanup method
func TempDir(t *testing.T) (out string, cleanup func()) {
	out, err := ioutil.TempDir("", strings.ReplaceAll(t.Name(), "/", "_")+"_")
	require.NoError(t, err)

	cleanup = func() {
		// for debug purpose we keep everything if the test failed
		if err != nil {
			require.NoError(t, os.RemoveAll(out))
		}
	}
	return out, cleanup
}
