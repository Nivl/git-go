package smoke_test

import (
	"testing"

	"github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestWorkingOnExistingRepo(t *testing.T) {
	t.Parallel()

	repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
	t.Cleanup(cleanup)

	// Create a new repo
	r, err := git.OpenRepository(repoPath)
	require.NoError(t, err, "failed opening a repo")
	t.Cleanup(func() {
		require.NoError(t, r.Close(), "failed closing repo")
	})

	defaultBranchName := ginternals.LocalBranchFullName("master")
	defaultBranch, err := r.GetReference(defaultBranchName)
	require.NoError(t, err, "couldn't get the default branch")

	// Update repo's readme
	headCommit, err := r.GetCommit(defaultBranch.Target())
	require.NoError(t, err, "couldn't get the head commit")
	rootTree, err := r.GetTree(headCommit.TreeID())
	require.NoError(t, err, "couldn't get the head commit's tree")
	rootTree.Entries()

	// Let's find the readme
	// TODO(melvin): Add a convenience method to find a file in a tree
	entries := rootTree.Entries()
	readmeOid := ginternals.NullOid
	for _, entry := range entries {
		if entry.Path == "README.md" {
			readmeOid = entry.ID
			break
		}
	}
	if readmeOid.IsZero() {
		t.Fatal("couldn't find the readme in the tree")
	}
	// TODO(melvin): Add a convenience method to get a blob
	readmeObj, err := r.GetObject(readmeOid)
	require.NoError(t, err, "failed finding the readme object from it's oid")
	readme := readmeObj.AsBlob()

	tb := r.NewTreeBuilderFromTree(rootTree)
	newReadme, err := r.NewBlob(append(readme.BytesCopy(), []byte("\nHello World\n")...))
	require.NoError(t, err, "failed creating new readme")
	err = tb.Insert("README.md", newReadme.ID(), object.ModeFile)
	require.NoError(t, err, "failed adding readme to tree")

	newTree, err := tb.Write()
	require.NoError(t, err, "failed creating new tree")

	fixBranchName := ginternals.LocalBranchFullName("ml/docs/update-readme")
	fixCommit, err := r.NewCommit(
		fixBranchName,
		newTree,
		object.NewSignature("John Doe", "john@domain.tld"),
		&object.CommitOptions{
			Message:   "docs(readme): Fix typo",
			ParentsID: []ginternals.Oid{headCommit.ID()},
		})
	require.NoError(t, err, "failed creating the commit with the updated readme")

	// TODO(melvin): Write the commit to packfile + push it to the remote

	// Alright, time to merge this new branch into the default one!

	mergeCommit, err := r.NewCommit(
		defaultBranchName,
		newTree,
		object.NewSignature("John Doe", "john@domain.tld"),
		&object.CommitOptions{
			Message:   "merge branch ml/docs/fix-typo-in-readme into main",
			ParentsID: []ginternals.Oid{headCommit.ID(), fixCommit.ID()},
		})
	require.NoError(t, err, "failed creating the commit with the fix")

	// Make sure the merge worked
	mainBranch, err := r.GetReference(defaultBranchName)
	require.NoError(t, err, "couldn't get the main branch")
	require.Equal(t, mergeCommit.ID(), mainBranch.Target(), "the merge didn't work")
}
