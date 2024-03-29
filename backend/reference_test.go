package backend

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/Nivl/git-go/internal/testutil/confutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReference(t *testing.T) {
	t.Parallel()

	t.Run("Should fail if reference doesn't exists", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		ref, err := b.Reference("refs/heads/doesnt_exists")
		require.Error(t, err)
		assert.True(t, errors.Is(ginternals.ErrRefNotFound, ginternals.ErrRefNotFound), "unexpected error returned")
		assert.Nil(t, ref)
	})

	t.Run("Should success to follow a symbolic ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		ref, err := b.Reference(ginternals.Head)
		require.NoError(t, err)
		require.NotNil(t, ref)

		expectedTarget, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		assert.Equal(t, ginternals.Head, ref.Name())
		assert.Equal(t, "refs/heads/ml/packfile/tests", ref.SymbolicTarget())
		assert.Equal(t, expectedTarget, ref.Target())
	})

	t.Run("Should success to follow an oid ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		ref, err := b.Reference(ginternals.LocalBranchFullName(ginternals.Master))
		require.NoError(t, err)
		require.NotNil(t, ref)

		expectedTarget, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		assert.Equal(t, ginternals.LocalBranchFullName(ginternals.Master), ref.Name())
		assert.Empty(t, ref.SymbolicTarget())
		assert.Equal(t, expectedTarget, ref.Target())
	})
}

func TestParsePackedRefs(t *testing.T) {
	t.Parallel()

	createRepo := func(t *testing.T) (dir string, cleanup func()) {
		t.Helper()

		dir, cleanup = testutil.TempDir(t)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)

		defer require.NoError(t, b.Close())
		require.NoError(t, b.Init(ginternals.Master))
		return dir, cleanup
	}

	t.Run("Should return empty list if no files", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := createRepo(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		count := 0
		b.refs.Range(func(key, value interface{}) bool {
			count++
			return true
		})
		// By default it should only have HEAD
		assert.Equal(t, 1, count, "invalid amount of refs")
	})

	t.Run("Should fail if file contains invalid data", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := createRepo(t)
		t.Cleanup(cleanup)

		fPath := filepath.Join(dir, ".git", "packed-refs")
		err := os.WriteFile(fPath, []byte("not valid data"), 0o644)
		require.NoError(t, err)

		cfg := confutil.NewCommonConfig(t, dir)
		_, err = NewFS(cfg)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ginternals.ErrPackedRefInvalid), "unexpected error received")
	})

	t.Run("Should pass with comments and annotations", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := createRepo(t)
		t.Cleanup(cleanup)

		fPath := filepath.Join(dir, ".git", "packed-refs")
		err := os.WriteFile(fPath, []byte("^de111c003b5661db802f17ac69419dcb9f4f3137\n# this is a comment"), 0o644)
		require.NoError(t, err)

		cfg := confutil.NewCommonConfig(t, dir)
		_, err = NewFS(cfg)
		require.NoError(t, err)
	})

	t.Run("Should correctly extract data", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		require.NoError(t, err)
		expected := map[string][]byte{
			"HEAD":                                  []byte("ref: refs/heads/ml/packfile/tests\n"),
			"ORIG_HEAD":                             []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089\n"),
			"refs/heads/master":                     []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089"),
			"refs/heads/ml/cleanup-062020":          []byte("b328320060eb503cf337c7cff281712ef236963a"),
			"refs/heads/ml/packfile/tests":          []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089"),
			"refs/heads/ml/tests":                   []byte("f0f70144f38695250606b86a50cff2b440a417f3"),
			"refs/remotes/origin/master":            []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089"),
			"refs/remotes/origin/ml/cleanup-062020": []byte("b328320060eb503cf337c7cff281712ef236963a"),
			"refs/remotes/origin/ml/feat/clone":     []byte("5f35f2dc6cec7356da02ca26192ce2bc3f271e79"),
			"refs/remotes/origin/HEAD":              []byte("ref: refs/remotes/origin/master\n"),
			"refs/stash":                            []byte("3fe6cf63fceced491a79fe634eb1e2c888225707"),
			"refs/tags/annotated":                   []byte("80316e01dbfdf5c2a8a20de66c747ecd4c4bd442\n"),
			"refs/tags/lightweight":                 []byte("bbb720a96e4c29b9950a4c577c98470a4d5dd089\n"),
		}

		count := 0
		b.refs.Range(func(key, value interface{}) bool {
			count++

			name := key.(string)
			expectation, ok := expected[name]
			assert.True(t, ok, "%s is missing in map", name)
			assert.Equal(t, string(expectation), string(value.([]byte)), "invalid value for key %s", name)
			return true
		})
		require.Equal(t, len(expected), count, "invalid amount of refs")
	})
}

func TestWriteReference(t *testing.T) {
	t.Parallel()

	t.Run("should pass writing a new symbolic reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init(ginternals.Master))

		ref := ginternals.NewSymbolicReference("HEAD", "refs/heads/master")
		err = b.WriteReference(ref)
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(b.Path(), "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/master\n", string(data))
	})

	t.Run("should pass writing a new oid reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init(ginternals.Master))

		target, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		ref := ginternals.NewReference("HEAD", target)
		err = b.WriteReference(ref)
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(b.Path(), "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, target.String()+"\n", string(data))
	})

	t.Run("should fail with invalid name", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init(ginternals.Master))

		ref := ginternals.NewSymbolicReference("H EAD", "refs/heads/master")
		err = b.WriteReference(ref)
		require.Error(t, err)
		require.True(t, errors.Is(err, ginternals.ErrRefNameInvalid), "unexpected error")
	})

	t.Run("should pass overwriting a symbolic reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		// assert current data on disk
		data, err := os.ReadFile(filepath.Join(b.Path(), "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/ml/packfile/tests\n", string(data))

		ref := ginternals.NewSymbolicReference("HEAD", "refs/heads/master")
		err = b.WriteReference(ref)
		require.NoError(t, err)

		data, err = os.ReadFile(filepath.Join(b.Path(), "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/master\n", string(data))
	})

	t.Run("should pass overwriting an oid reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		// assert current data on disk
		data, err := os.ReadFile(filepath.Join(b.Path(), "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/ml/packfile/tests\n", string(data))

		target, err := ginternals.NewOidFromStr("abb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		ref := ginternals.NewReference("HEAD", target)
		err = b.WriteReference(ref)
		require.NoError(t, err)

		data, err = os.ReadFile(filepath.Join(b.Path(), "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, target.String()+"\n", string(data))
	})

	t.Run("should pass writing a reference containing '/'", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init(ginternals.Master))

		target, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		ref := ginternals.NewReference("ml/tests/references", target)
		err = b.WriteReference(ref)
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(b.Path(), "ml", "tests", "references"))
		require.NoError(t, err)
		assert.Equal(t, target.String()+"\n", string(data))
	})

	t.Run("should fail writing a reference containing '/' already used by another reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init(ginternals.Master))

		target, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		ref := ginternals.NewReference("ml/tests", target)
		err = b.WriteReference(ref)
		require.NoError(t, err)

		ref = ginternals.NewReference("ml/tests/references", target)
		err = b.WriteReference(ref)
		require.Error(t, err)
		// TODO(melvin): check error type. Windows doesn't fail on the MkdirAll
		// Making it hard to have a cross-platform test right now.
		// require.Contains(t, err.Error(), "not a directory")
	})

	t.Run("validate name", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		testCases := []struct {
			name        string
			expectError bool
		}{
			{
				name:        "refs/heads/master/2",
				expectError: true,
			},
			{
				name:        "refs/heads",
				expectError: true,
			},
			{
				name:        "refs/heads/master2",
				expectError: false,
			},
			{
				name:        "refs/heads2",
				expectError: false,
			},
			{
				name:        "refs/heads/master",
				expectError: false,
			},
		}
		for i, tc := range testCases {
			tc := tc
			i := i
			t.Run(fmt.Sprintf("%d/%s", i, tc.name), func(t *testing.T) {
				t.Parallel()

				ref := ginternals.NewSymbolicReference(tc.name, "refs/heads/master")
				err := b.WriteReference(ref)
				if tc.expectError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}

func TestWriteReferenceSafe(t *testing.T) {
	t.Parallel()

	t.Run("should pass writing a new symbolic reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init(ginternals.Master))

		ref := ginternals.NewSymbolicReference("refs/heads/my_feature", "refs/heads/master")
		err = b.WriteReferenceSafe(ref)
		require.NoError(t, err)

		// Let's make sure the data changed on disk
		data, err := os.ReadFile(filepath.Join(b.Path(), "refs", "heads", "my_feature"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/master\n", string(data))
	})

	t.Run("should pass writing a new oid reference", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init(ginternals.Master))

		target, err := ginternals.NewOidFromStr("bbb720a96e4c29b9950a4c577c98470a4d5dd089")
		require.NoError(t, err)
		ref := ginternals.NewReference("refs/heads/my_feature", target)
		err = b.WriteReferenceSafe(ref)
		require.NoError(t, err)

		// Let's make sure the data changed on disk
		data, err := os.ReadFile(filepath.Join(b.Path(), "refs", "heads", "my_feature"))
		require.NoError(t, err)
		assert.Equal(t, target.String()+"\n", string(data))
	})

	t.Run("should fail with invalid name", func(t *testing.T) {
		t.Parallel()

		dir, cleanup := testutil.TempDir(t)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, dir)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})
		require.NoError(t, b.Init(ginternals.Master))

		ref := ginternals.NewSymbolicReference("H EAD", "refs/heads/master")
		err = b.WriteReferenceSafe(ref)
		require.Error(t, err)
		require.True(t, errors.Is(err, ginternals.ErrRefNameInvalid), "unexpected error")
	})

	t.Run("should fail overwritting a ref on disk", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		// assert current data on disk
		data, err := os.ReadFile(filepath.Join(b.Path(), "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/ml/packfile/tests\n", string(data))

		ref := ginternals.NewSymbolicReference("HEAD", "refs/heads/master")
		err = b.WriteReferenceSafe(ref)
		require.Error(t, err)
		require.True(t, errors.Is(err, ginternals.ErrRefExists), "unexpected error")

		// let's make sure the data have not changed
		data, err = os.ReadFile(filepath.Join(b.Path(), "HEAD"))
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/ml/packfile/tests\n", string(data))
	})

	t.Run("should fail overwritting a packed ref", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		// assert current data on disk (there are none)
		_, err = os.ReadFile(filepath.Join(b.Path(), "refs", "heads", "master"))
		require.Error(t, err)

		ref := ginternals.NewSymbolicReference("refs/heads/master", "refs/heads/branch")
		err = b.WriteReferenceSafe(ref)
		require.Error(t, err)
		require.True(t, errors.Is(err, ginternals.ErrRefExists), "unexpected error")

		// Let's make sure the data have not been persisted
		_, err = os.ReadFile(filepath.Join(b.Path(), "refs", "heads", "master"))
		require.Error(t, err)
	})
}

func TestWalkReferences(t *testing.T) {
	t.Parallel()

	t.Run("should pass writing a new symbolic reference", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		var count int
		err = b.WalkReferences(func(ref *ginternals.Reference) error {
			count++
			return nil
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 10)
	})

	t.Run("should stop with WalkStop", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		var count int
		err = b.WalkReferences(func(ref *ginternals.Reference) error {
			if count == 4 {
				return WalkStop
			}
			count++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 4, count)
	})

	t.Run("should bubble up the provided error", func(t *testing.T) {
		t.Parallel()

		repoPath, cleanup := testutil.UnTar(t, testutil.RepoSmall)
		t.Cleanup(cleanup)

		cfg := confutil.NewCommonConfig(t, repoPath)
		b, err := NewFS(cfg)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, b.Close())
		})

		someError := errors.New("some error")
		var count int
		err = b.WalkReferences(func(ref *ginternals.Reference) error {
			if count == 4 {
				return someError
			}
			count++
			return nil
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, someError)
	})
}
