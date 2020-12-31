package cache

import (
	"sync"

	lru "github.com/golang/groupcache/lru"
)

// LRUKey may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type LRUKey = lru.Key

// LRU represents a LRU cache
type LRU struct {
	cache *lru.Cache
	mu    sync.Mutex
}

// NewLRU creates a new LRU Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func NewLRU(maxEntries int) *LRU {
	return &LRU{
		cache: lru.New(maxEntries),
	}
}

// Get looks up a key's value from the cache.
func (c *LRU) Get(key LRUKey) (value interface{}, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cache.Get(key)
}

// Add adds a value to the cache.
func (c *LRU) Add(key LRUKey, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Add(key, value)
}

// Clear purges all stored items from the cache.
func (c *LRU) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Clear()
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cache.Len()
}
