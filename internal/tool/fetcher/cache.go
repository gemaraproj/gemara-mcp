// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"sync"
	"time"
)

// Cache is a generic, thread-safe TTL cache keyed by string.
type Cache[T any] struct {
	mu    sync.RWMutex
	items map[string]cacheItem[T]
	ttl   time.Duration
}

type cacheItem[T any] struct {
	value     T
	source    string
	cacheTime time.Time
}

// NewCache creates a new cache with the specified TTL.
func NewCache[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		items: make(map[string]cacheItem[T]),
		ttl:   ttl,
	}
}

// Get retrieves a cached value if available and not expired.
func (c *Cache[T]) Get(key string) (T, string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		var zero T
		return zero, "", false
	}

	if time.Since(item.cacheTime) >= c.ttl {
		var zero T
		return zero, "", false
	}

	return item.value, item.source, true
}

// Set stores a value in the cache.
func (c *Cache[T]) Set(key string, value T, source string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem[T]{
		value:     value,
		source:    source,
		cacheTime: time.Now(),
	}
}
