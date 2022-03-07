package smoke_test

import (
	"testing"

	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestWorkingOnNewRepo(t *testing.T) {
	t.Parallel()

	d, cleanup := testutil.TempDir(t)
	t.Cleanup(cleanup)

	// Create a new repo
	r, err := git.InitRepository(d)
	require.NoError(t, err, "failed creating a repo")
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	// Add new files to the repo
	tb := r.NewTreeBuilder()

	readme, err := r.NewBlob([]byte("Hello Wrld\n"))
	require.NoError(t, err, "failed creating readme")
	err = tb.Insert("README.md", readme.ID(), object.ModeFile)
	require.NoError(t, err, "failed adding readme to tree")

	rootTree, err := tb.Write()
	require.NoError(t, err, "failed creating root tree")

	// Create the first commit in the main branch
	mainBranchName := ginternals.LocalBranchFullName("main")
	initialCommit, err := r.NewCommit(
		mainBranchName,
		rootTree,
		object.NewSignature("John Doe", "john@domain.tld"),
		&object.CommitOptions{
			Message: "Initial commit",
		})
	require.NoError(t, err, "failed creating the initial commit")

	// TODO(melvin): Write the commit to packfile + push it to the remote

	// Oops, we have a typo in our Readme. Let's fix it by creating a new
	// commit in another branch and merging them.

	tb = r.NewTreeBuilderFromTree(rootTree)

	readme, err = r.NewBlob([]byte("Hello World\n"))
	require.NoError(t, err, "failed creating 2nd version of readme")
	err = tb.Insert("README.md", readme.ID(), object.ModeFile)
	require.NoError(t, err, "failed updating readme in tree")

	newTree, err := tb.Write()
	require.NoError(t, err, "failed creating new tree")

	fixBranchName := ginternals.LocalBranchFullName("ml/docs/fix-typo-in-readme")
	fixCommit, err := r.NewCommit(
		fixBranchName,
		newTree,
		object.NewSignature("John Doe", "john@domain.tld"),
		&object.CommitOptions{
			Message:   "docs(readme): Fix typo",
			ParentsID: []ginternals.Oid{initialCommit.ID()},
		})
	require.NoError(t, err, "failed creating the commit with the fix")

	// TODO(melvin): Write the commit to packfile + push it to the remote

	// Alright, time to merge this new branch into the default one!

	mergeCommit, err := r.NewCommit(
		mainBranchName,
		newTree,
		object.NewSignature("John Doe", "john@domain.tld"),
		&object.CommitOptions{
			Message:   "merge branch ml/docs/fix-typo-in-readme into main",
			ParentsID: []ginternals.Oid{initialCommit.ID(), fixCommit.ID()},
		})
	require.NoError(t, err, "failed creating the commit with the fix")

	// Make sure the merge worked
	mainBranch, err := r.Reference(mainBranchName)
	require.NoError(t, err, "couldn't get the main branch")
	require.Equal(t, mergeCommit.ID(), mainBranch.Target(), "the merge didn't work")
}
