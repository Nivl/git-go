package main

import (
	"fmt"
	"testing"

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
