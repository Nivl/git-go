package ginternals

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Nivl/git-go/ginternals/githash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRefNameValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc       string
		name       string
		shouldPass bool
	}{
		{
			desc:       "name with control chars should fail",
			name:       "ml/not\000valide",
			shouldPass: false,
		},
		{
			desc:       "name with control chars should fail",
			name:       "ml/not\177valide",
			shouldPass: false,
		},
		{
			desc:       "name with slashes should pass",
			name:       "ml/some/name_/that/I/often-use/89",
			shouldPass: true,
		},
		{
			desc:       "name cannot be empty",
			name:       "",
			shouldPass: false,
		},
		{
			desc:       "name cannot start with a /",
			name:       "/refs/heads/master",
			shouldPass: false,
		},
		{
			desc:       "name cannot end with a /",
			name:       "refs/heads/master/",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain ..",
			name:       "refs/heads/ma..ster",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain ?",
			name:       "refs/heads/master?",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain :",
			name:       "refs/heads/ma:ster",
			shouldPass: false,
		},
		{
			desc:       `name cannot contain \`,
			name:       `refs/heads/ma\ster`,
			shouldPass: false,
		},
		{
			desc:       "name cannot contain ^",
			name:       "refs/heads/ma^ster",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain @{",
			name:       "refs/heads/ma@{ster}",
			shouldPass: false,
		},
		{
			desc:       "name can end with @",
			name:       "refs/heads/master@",
			shouldPass: true,
		},
		{
			desc:       "name cannot start with a .",
			name:       ".refs/heads/master",
			shouldPass: false,
		},
		{
			desc:       "name cannot end with a .",
			name:       "refs/heads/master.",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain a [",
			name:       "[refs/heads/master",
			shouldPass: false,
		},
		{
			desc:       "name cannot contain a space",
			name:       "refs/he ads/master",
			shouldPass: false,
		},
		{
			desc:       "name cannot end with .lock",
			name:       "refs/heads/master.lock",
			shouldPass: false,
		},
		{
			desc:       "segments cannot be empty",
			name:       "refs//master",
			shouldPass: false,
		},
		{
			desc:       "segments cannot end with a .",
			name:       "refs/heads./master",
			shouldPass: false,
		},
		{
			desc:       "segments cannot end with .lock",
			name:       "refs/heads.lock/master",
			shouldPass: false,
		},
		{
			desc:       "HEAD should be a valid reference",
			name:       "HEAD",
			shouldPass: true,
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()

			res := IsRefNameValid(tc.name)
			assert.Equal(t, tc.shouldPass, res)
		})
	}
}

func TestResolveReference(t *testing.T) {
	t.Parallel()

	t.Run("should resolve oid reference", func(t *testing.T) {
		t.Parallel()

		finder := func(name string) ([]byte, error) {
			switch name {
			case "refs/heads/master":
				return []byte("0eaf966ff79d8f61958aaefe163620d952606516\n"), nil
			default:
				return nil, errors.New("unexpected")
			}
		}
		hash := githash.NewSHA1()
		ref, err := ResolveReference(hash, "refs/heads/master", finder)
		require.NoError(t, err)
		assert.Equal(t, OidReference, ref.Type())
		assert.Empty(t, ref.SymbolicTarget())
		assert.Equal(t, "0eaf966ff79d8f61958aaefe163620d952606516", ref.Target().String())
	})

	t.Run("should resolve symbolic reference", func(t *testing.T) {
		t.Parallel()

		finder := func(name string) ([]byte, error) {
			switch name {
			case "HEAD":
				return []byte("ref: refs/heads/master\n"), nil
			case "refs/heads/master":
				return []byte("0eaf966ff79d8f61958aaefe163620d952606516\n"), nil
			default:
				return nil, errors.New("unexpected")
			}
		}
		hash := githash.NewSHA1()
		ref, err := ResolveReference(hash, "HEAD", finder)
		require.NoError(t, err)
		assert.Equal(t, SymbolicReference, ref.Type())
		assert.Equal(t, "refs/heads/master", ref.SymbolicTarget())
		assert.Equal(t, "0eaf966ff79d8f61958aaefe163620d952606516", ref.Target().String())
	})

	t.Run("should fail on loops", func(t *testing.T) {
		t.Parallel()

		finder := func(name string) ([]byte, error) {
			switch name {
			case "HEAD":
				return []byte("ref: refs/heads/master\n"), nil
			case "refs/heads/master":
				return []byte("ref: HEAD\n"), nil
			default:
				return nil, errors.New("unexpected")
			}
		}
		hash := githash.NewSHA1()
		ref, err := ResolveReference(hash, "HEAD", finder)
		require.Error(t, err)
		assert.Nil(t, ref)
		assert.True(t, errors.Is(err, ErrRefInvalid), "invalid error returned")
	})

	t.Run("should fail op invalid name", func(t *testing.T) {
		t.Parallel()

		finder := func(name string) ([]byte, error) {
			switch name {
			case "HEAD":
				return []byte("ref: refs/hea ds/master\n"), nil
			default:
				return nil, errors.New("unexpected")
			}
		}
		hash := githash.NewSHA1()
		ref, err := ResolveReference(hash, "HEAD", finder)
		require.Error(t, err)
		assert.Nil(t, ref)
		assert.True(t, errors.Is(err, ErrRefNameInvalid), "invalid error returned")
	})

	t.Run("should fail on invalid content", func(t *testing.T) {
		t.Parallel()

		finder := func(name string) ([]byte, error) {
			switch name {
			case "HEAD":
				return []byte("not a valid ref\n"), nil
			default:
				return nil, errors.New("unexpected")
			}
		}
		hash := githash.NewSHA1()
		ref, err := ResolveReference(hash, "HEAD", finder)
		require.Error(t, err)
		assert.Nil(t, ref)
		assert.True(t, errors.Is(err, ErrRefInvalid), "invalid error returned")
	})

	t.Run("should fail on empty file", func(t *testing.T) {
		t.Parallel()

		finder := func(name string) ([]byte, error) {
			switch name {
			case "HEAD":
				return []byte(""), nil
			default:
				return nil, errors.New("unexpected")
			}
		}
		hash := githash.NewSHA1()
		ref, err := ResolveReference(hash, "HEAD", finder)
		require.Error(t, err)
		assert.Nil(t, ref)
		assert.True(t, errors.Is(err, ErrRefInvalid), "invalid error returned")
	})

	t.Run("should pass error down from the finder", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		finder := func(name string) ([]byte, error) {
			return nil, expectedErr
		}
		hash := githash.NewSHA1()
		ref, err := ResolveReference(hash, "HEAD", finder)
		require.Error(t, err)
		assert.Nil(t, ref)
		assert.True(t, errors.Is(err, expectedErr), "invalid error returned")
	})
}

func TestNewReference(t *testing.T) {
	t.Parallel()

	hash := githash.NewSHA1()
	oid, err := hash.ConvertFromString("0eaf966ff79d8f61958aaefe163620d952606516")
	require.NoError(t, err)

	ref := NewReference("HEAD", oid)
	assert.Equal(t, OidReference, ref.Type())
	assert.Equal(t, "HEAD", ref.Name())
	assert.Empty(t, ref.SymbolicTarget())
	assert.Equal(t, "0eaf966ff79d8f61958aaefe163620d952606516", ref.Target().String())
}

func TestNewSymbolicReference(t *testing.T) {
	t.Parallel()

	hash := githash.NewSHA1()
	ref := NewSymbolicReference("HEAD", "refs/heads/master")
	assert.Equal(t, SymbolicReference, ref.Type())
	assert.Equal(t, "HEAD", ref.Name())
	assert.Equal(t, hash.NullOid(), ref.Target())
	assert.Equal(t, "refs/heads/master", ref.SymbolicTarget())
}
