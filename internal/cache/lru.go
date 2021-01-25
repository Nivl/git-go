package cache

import (
	"sync"

	lru "github.com/hashicorp/golang-lru"
)

// LRU represents a LRU cache
type LRU struct {
	cache *lru.Cache
	mu    sync.Mutex
}

// NewLRU creates a new LRU Cache.
func NewLRU(maxEntries int) (*LRU, error) {
	cache, err := lru.New(maxEntries)
	if err != nil {
		return nil, err
	}
	return &LRU{
		cache: cache,
	}, nil
}

// Get looks up a key's value from the cache.
func (c *LRU) Get(key interface{}) (value interface{}, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cache.Get(key)
}

// Add adds a value to the cache.
func (c *LRU) Add(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Add(key, value)
}

// Clear purges all stored items from the cache.
func (c *LRU) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Purge()
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cache.Len()
}
