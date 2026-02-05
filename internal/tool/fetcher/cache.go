// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"sync"
	"time"
)

// TODO(jpower432): We can probably use a library here, but because we only need something really
// simple right now - implemented it directly.

// Cache is a shared cache for raw bytes fetched by fetchers, keyed by source identifier.
type Cache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
	ttl   time.Duration
}

type cacheItem struct {
	data      []byte
	source    string
	cacheTime time.Time
}

// NewCache creates a new shared fetcher cache with the specified TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]cacheItem),
		ttl:   ttl,
	}
}

// Get retrieves cached data for a source if available and not expired.
func (c *Cache) Get(source string) ([]byte, string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[source]
	if !found {
		return nil, "", false
	}

	// Check if expired
	if time.Since(item.cacheTime) >= c.ttl {
		return nil, "", false
	}

	return item.data, item.source, true
}

// Set stores data in the cache for a source.
func (c *Cache) Set(source string, data []byte, sourceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[source] = cacheItem{
		data:      data,
		source:    sourceID,
		cacheTime: time.Now(),
	}
}
