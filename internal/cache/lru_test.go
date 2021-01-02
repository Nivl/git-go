package cache_test

import (
	"testing"

	"github.com/Nivl/git-go/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestLRU(t *testing.T) {
	t.Parallel()

	t.Run("Add and get data", func(t *testing.T) {
		t.Parallel()

		c := cache.NewLRU(1)

		assert.Equal(t, 0, c.Len(), "expected an empty cache")

		rv, ok := c.Get("key")
		assert.False(t, ok, "should not find data that does not exist")
		assert.Nil(t, rv, "returned value should be nil when not found")

		c.Add("key", 1)
		assert.Equal(t, 1, c.Len(), "expected 1 item in the cache")

		var v int
		rv, ok = c.Get("key")
		assert.True(t, ok, "should have found data")
		assert.NotPanics(t, func() {
			v = rv.(int)
		})
		assert.Equal(t, 1, v, "unexpected data retrieved from cache")

		c.Clear()
		assert.Equal(t, 0, c.Len(), "expected the cache t have been emptied")
	})
}
