// Package testhelper contains helpers to simplify tests
package testhelper

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TempDir creates a temp dir and returns a cleanup method
func TempDir(t *testing.T) (out string, cleanup func()) {
	out, err := os.MkdirTemp("", strings.ReplaceAll(t.Name(), "/", "_")+"_")
	require.NoError(t, err)

	cleanup = func() {
		require.NoError(t, os.RemoveAll(out))
	}
	return out, cleanup
}

// TempFile creates a temp file and returns a cleanup method
func TempFile(t *testing.T) (out *os.File, cleanup func()) {
	out, err := os.CreateTemp("", strings.ReplaceAll(t.Name(), "/", "_")+"_")
	require.NoError(t, err)

	cleanup = func() {
		require.NoError(t, out.Close())
		require.NoError(t, os.RemoveAll(out.Name()))
	}
	return out, cleanup
}
