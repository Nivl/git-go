package env

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFromOs(t *testing.T) {
	t.Parallel()

	e := NewFromOs()
	// 5 here is a random use value, it's assuming than the system
	// running the tests has more than 5 variable
	assert.Greater(t, len(e.env), 5)
}

func TestNewFromKVList(t *testing.T) {
	t.Parallel()

	e := NewFromKVList([]string{
		"VERSION=1",
		"ENABLE=true",
		"PATH=a:b:c",
		"X=",
	})
	assert.Len(t, e.env, 4)
	assert.Equal(t, map[string]string{
		"VERSION": "1",
		"ENABLE":  "true",
		"PATH":    "a:b:c",
		"X":       "",
	}, e.env)
}

func TestGet(t *testing.T) {
	t.Parallel()

	e := NewFromKVList([]string{
		"VERSION=1",
		"ENABLE=true",
		"PATH=a:b:c",
		"X=",
	})

	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "existing key",
			input:    "VERSION",
			expected: "1",
		},
		{
			desc:     "existing key invalid case",
			input:    "version",
			expected: "",
		},
		{
			desc:     "non existing key",
			input:    "nope",
			expected: "",
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, e.Get(tc.input))
		})
	}
}

func TestHas(t *testing.T) {
	t.Parallel()

	e := NewFromKVList([]string{
		"VERSION=1",
		"ENABLE=true",
		"PATH=a:b:c",
		"X=",
	})

	testCases := []struct {
		desc     string
		input    string
		expected bool
	}{
		{
			desc:     "existing key",
			input:    "VERSION",
			expected: true,
		},
		{
			desc:     "existing key invalid case",
			input:    "version",
			expected: false,
		},
		{
			desc:     "non existing key",
			input:    "nope",
			expected: false,
		},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, e.Has(tc.input))
		})
	}
}
