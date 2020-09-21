package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nivl/git-go/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatFileParams(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		args []string
	}{
		{
			desc: "-t cannot be used with -p",
			args: []string{"cat-file", "-p", "-t", "642480605b8b0fd464ab5762e044269cf29a60a3"},
		},
		{
			desc: "-s cannot be used with -p",
			args: []string{"cat-file", "-p", "-s", "642480605b8b0fd464ab5762e044269cf29a60a3"},
		},
		{
			desc: "-s cannot be used with -t",
			args: []string{"cat-file", "-t", "-s", "642480605b8b0fd464ab5762e044269cf29a60a3"},
		},
		{
			desc: "no type allowed with -t",
			args: []string{"cat-file", "-t", "blob", "642480605b8b0fd464ab5762e044269cf29a60a3"},
		},
		{
			desc: "no type allowed with -s",
			args: []string{"cat-file", "-s", "blob", "642480605b8b0fd464ab5762e044269cf29a60a3"},
		},
		{
			desc: "no type allowed with -p",
			args: []string{"cat-file", "-p", "blob", "642480605b8b0fd464ab5762e044269cf29a60a3"},
		},
		{
			desc: "type required when no -p -s -t",
			args: []string{"cat-file", "642480605b8b0fd464ab5762e044269cf29a60a3"},
		},
		{
			desc: "sha required when no -p -s -t",
			args: []string{"cat-file", "blob"},
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			cmd := newRootCmd()
			cmd.SetArgs(tc.args)

			var err error
			require.NotPanics(t, func() {
				err = cmd.Execute()
			})
			require.Error(t, err)
		})
	}
}

func TestCatFile(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testhelper.UnTar(t, testhelper.RepoSmall)
	t.Cleanup(cleanup)

	testCases := []struct {
		desc           string
		args           []string
		expectedOutput string
	}{
		{
			desc:           "-s should print the size (blob)",
			args:           []string{"cat-file", "-s", "642480605b8b0fd464ab5762e044269cf29a60a3"},
			expectedOutput: "453\n",
		},
		{
			desc:           "-t should print the type (blob)",
			args:           []string{"cat-file", "-t", "642480605b8b0fd464ab5762e044269cf29a60a3"},
			expectedOutput: "blob\n",
		},
		{
			desc:           "-p should pretty-print (blob)",
			args:           []string{"cat-file", "-p", "642480605b8b0fd464ab5762e044269cf29a60a3"},
			expectedOutput: "file://blob_642480605b8b0fd464ab5762e044269cf29a60a3",
		},
		{
			desc:           "default should print raw object (blob)",
			args:           []string{"cat-file", "blob", "642480605b8b0fd464ab5762e044269cf29a60a3"},
			expectedOutput: "file://blob_642480605b8b0fd464ab5762e044269cf29a60a3",
		},
		{
			desc:           "-s should print the size (tree)",
			args:           []string{"cat-file", "-s", "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3"},
			expectedOutput: "463\n",
		},
		{
			desc:           "-t should print the type (tree)",
			args:           []string{"cat-file", "-t", "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3"},
			expectedOutput: "tree\n",
		},
		{
			desc:           "-p should pretty-print (tree)",
			args:           []string{"cat-file", "-p", "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3"},
			expectedOutput: "file://tree_e5b9e846e1b468bc9597ff95d71dfacda8bd54e3_pretty",
		},
		{
			desc:           "default should print raw object (tree)",
			args:           []string{"cat-file", "tree", "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3"},
			expectedOutput: "file://tree_e5b9e846e1b468bc9597ff95d71dfacda8bd54e3",
		},
		{
			desc:           "-s should print the size (commit)",
			args:           []string{"cat-file", "-s", "bbb720a96e4c29b9950a4c577c98470a4d5dd089"},
			expectedOutput: "260\n",
		},
		{
			desc:           "-t should print the type (commit)",
			args:           []string{"cat-file", "-t", "bbb720a96e4c29b9950a4c577c98470a4d5dd089"},
			expectedOutput: "commit\n",
		},
		{
			desc:           "-p should pretty-print (commit)",
			args:           []string{"cat-file", "-p", "bbb720a96e4c29b9950a4c577c98470a4d5dd089"},
			expectedOutput: "file://commit_bbb720a96e4c29b9950a4c577c98470a4d5dd089",
		},
		{
			desc:           "default should print raw object (commit)",
			args:           []string{"cat-file", "commit", "bbb720a96e4c29b9950a4c577c98470a4d5dd089"},
			expectedOutput: "file://commit_bbb720a96e4c29b9950a4c577c98470a4d5dd089",
		},
		{
			desc:           "default should print raw object (annotated tag)",
			args:           []string{"cat-file", "-p", "annotated"},
			expectedOutput: "file://annotated",
		},
		{
			desc:           "default should print raw object (HEAD)",
			args:           []string{"cat-file", "-p", "HEAD"},
			expectedOutput: "file://commit_bbb720a96e4c29b9950a4c577c98470a4d5dd089",
		},
		{
			desc:           "default should print raw object (refs/heads/ml/packfile/tests)",
			args:           []string{"cat-file", "-p", "refs/heads/ml/packfile/tests"},
			expectedOutput: "file://commit_bbb720a96e4c29b9950a4c577c98470a4d5dd089",
		},
		{
			desc:           "default should print raw object (heads/ml/packfile/tests)",
			args:           []string{"cat-file", "-p", "heads/ml/packfile/tests"},
			expectedOutput: "file://commit_bbb720a96e4c29b9950a4c577c98470a4d5dd089",
		},
		{
			desc:           "default should print raw object (ml/packfile/tests)",
			args:           []string{"cat-file", "-p", "ml/packfile/tests"},
			expectedOutput: "file://commit_bbb720a96e4c29b9950a4c577c98470a4d5dd089",
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			outBuf := bytes.NewBufferString("")
			cmd := newRootCmd()
			cmd.SetOut(outBuf)
			args := append([]string{"-C", repoPath}, tc.args...)
			cmd.SetArgs(args)

			var err error
			require.NotPanics(t, func() {
				err = cmd.Execute()
			})
			require.NoError(t, err)

			out, err := ioutil.ReadAll(outBuf)
			require.NoError(t, err)

			expected := tc.expectedOutput
			if strings.HasPrefix(tc.expectedOutput, "file://") {
				filename := strings.TrimPrefix(tc.expectedOutput, "file://")
				content, err := ioutil.ReadFile(filepath.Join(testhelper.TestdataPath(t), filename))
				require.NoError(t, err)
				expected = string(content)
			}
			assert.Equal(t, expected, string(out))
		})
	}
}
