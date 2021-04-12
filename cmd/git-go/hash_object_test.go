package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/env"
	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashObjectCmd(t *testing.T) {
	t.Parallel()

	t.Run("blob", func(t *testing.T) {
		t.Parallel()

		t.Run("default should be blob", func(t *testing.T) {
			t.Parallel()

			repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
			t.Cleanup(cleanup)

			cwd, err := os.Getwd()
			require.NoError(t, err)

			outBuf := bytes.NewBufferString("")
			cmd := newRootCmd(cwd, env.NewFromOs())
			cmd.SetArgs([]string{
				"hash-object",
				filepath.Join(repoPath, "README.md"),
			})
			cmd.SetOut(outBuf)

			require.NotPanics(t, func() {
				err = cmd.Execute()
			})
			require.NoError(t, err)
			out, err := io.ReadAll(outBuf)
			require.NoError(t, err)

			assert.Equal(t, "642480605b8b0fd464ab5762e044269cf29a60a3\n", string(out))
		})

		t.Run("blob opt should work", func(t *testing.T) {
			t.Parallel()

			cwd, err := os.Getwd()
			require.NoError(t, err)

			outBuf := bytes.NewBufferString("")
			cmd := newRootCmd(cwd, env.NewFromOs())
			require.NoError(t, err)
			cmd.SetArgs([]string{
				"hash-object",
				"-t", "blob",
				filepath.Join(testhelper.TestdataPath(t), "blob"),
			})
			cmd.SetOut(outBuf)

			require.NotPanics(t, func() {
				err = cmd.Execute()
			})
			require.NoError(t, err)
			out, err := io.ReadAll(outBuf)
			require.NoError(t, err)

			assert.Equal(t, "286db5050f814069644960e6cc7589c386053c6c\n", string(out))
		})
	})

	t.Run("tree", func(t *testing.T) {
		t.Parallel()

		t.Run("valid tree should work", func(t *testing.T) {
			t.Parallel()

			cwd, err := os.Getwd()
			require.NoError(t, err)

			outBuf := bytes.NewBufferString("")
			cmd := newRootCmd(cwd, env.NewFromOs())
			require.NoError(t, err)
			cmd.SetArgs([]string{
				"hash-object",
				"-t", "tree",
				filepath.Join(testhelper.TestdataPath(t), "tree"),
			})
			cmd.SetOut(outBuf)

			require.NotPanics(t, func() {
				err = cmd.Execute()
			})
			require.NoError(t, err)
			out, err := io.ReadAll(outBuf)
			require.NoError(t, err)

			assert.Equal(t, "2651fee5e238156738bc05ed1b558fdc9dc56fde\n", string(out))
		})

		t.Run("invalid tree should fail", func(t *testing.T) {
			t.Parallel()

			cwd, err := os.Getwd()
			require.NoError(t, err)

			outBuf := bytes.NewBufferString("")
			cmd := newRootCmd(cwd, env.NewFromOs())
			require.NoError(t, err)
			cmd.SetArgs([]string{
				"hash-object",
				"-t", "tree",
				filepath.Join(testhelper.TestdataPath(t), "blob"),
			})
			cmd.SetOut(outBuf)

			require.NotPanics(t, func() {
				err = cmd.Execute()
			})
			require.Error(t, err)

			// let's make sure we have mo content
			out, err := io.ReadAll(outBuf)
			require.NoError(t, err)
			assert.Empty(t, string(out))
		})
	})

	t.Run("commit", func(t *testing.T) {
		t.Parallel()

		t.Run("valid commit should work", func(t *testing.T) {
			t.Parallel()

			cwd, err := os.Getwd()
			require.NoError(t, err)

			outBuf := bytes.NewBufferString("")
			cmd := newRootCmd(cwd, env.NewFromOs())
			require.NoError(t, err)
			cmd.SetArgs([]string{
				"hash-object",
				"-t", "commit",
				filepath.Join(testhelper.TestdataPath(t), "commit"),
			})
			cmd.SetOut(outBuf)

			require.NotPanics(t, func() {
				err = cmd.Execute()
			})
			require.NoError(t, err)
			out, err := io.ReadAll(outBuf)
			require.NoError(t, err)

			assert.Equal(t, "0499018e26f79d37ad056611b75730dcb12918fb\n", string(out))
		})

		t.Run("invalid commit should fail", func(t *testing.T) {
			t.Parallel()

			cwd, err := os.Getwd()
			require.NoError(t, err)

			outBuf := bytes.NewBufferString("")
			cmd := newRootCmd(cwd, env.NewFromOs())
			require.NoError(t, err)
			cmd.SetArgs([]string{
				"hash-object",
				"-t", "commit",
				filepath.Join(testhelper.TestdataPath(t), "tree"),
			})
			cmd.SetOut(outBuf)

			require.NotPanics(t, func() {
				err = cmd.Execute()
			})
			assert.Error(t, err)

			// let's make sure we have mo content
			out, err := io.ReadAll(outBuf)
			require.NoError(t, err)
			assert.Empty(t, string(out))
		})
	})
}
