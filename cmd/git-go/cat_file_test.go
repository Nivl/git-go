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
			args:           []string{"cat-file", "-s", "2651fee5e238156738bc05ed1b558fdc9dc56fde"},
			expectedOutput: "178\n",
		},
		{
			desc:           "-t should print the type (tree)",
			args:           []string{"cat-file", "-t", "2651fee5e238156738bc05ed1b558fdc9dc56fde"},
			expectedOutput: "tree\n",
		},
		{
			desc:           "-p should pretty-print (tree)",
			args:           []string{"cat-file", "-p", "2651fee5e238156738bc05ed1b558fdc9dc56fde"},
			expectedOutput: "file://tree_2651fee5e238156738bc05ed1b558fdc9dc56fde_pretty",
		},
		{
			desc:           "default should print raw object (tree)",
			args:           []string{"cat-file", "tree", "2651fee5e238156738bc05ed1b558fdc9dc56fde"},
			expectedOutput: "file://tree_2651fee5e238156738bc05ed1b558fdc9dc56fde",
		},
		{
			desc:           "-s should print the size (commit)",
			args:           []string{"cat-file", "-s", "05986554a371782e3523cb98fcba5cd12fd90669"},
			expectedOutput: "733\n",
		},
		{
			desc:           "-t should print the type (commit)",
			args:           []string{"cat-file", "-t", "05986554a371782e3523cb98fcba5cd12fd90669"},
			expectedOutput: "commit\n",
		},
		{
			desc:           "-p should pretty-print (commit)",
			args:           []string{"cat-file", "-p", "05986554a371782e3523cb98fcba5cd12fd90669"},
			expectedOutput: "file://commit_05986554a371782e3523cb98fcba5cd12fd90669",
		},
		{
			desc:           "default should print raw object (commit)",
			args:           []string{"cat-file", "commit", "05986554a371782e3523cb98fcba5cd12fd90669"},
			expectedOutput: "file://commit_05986554a371782e3523cb98fcba5cd12fd90669",
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
			cmd.SetArgs(tc.args)

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
