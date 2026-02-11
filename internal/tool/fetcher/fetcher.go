// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher is a generic interface for fetching raw data from a source.
type Fetcher interface {
	// Fetch retrieves raw data from the source and returns the data and source identifier.
	Fetch(ctx context.Context) ([]byte, string, error)
}

// HTTPFetcher fetches data from an HTTP URL.
type HTTPFetcher struct {
	url     string
	timeout time.Duration
}

// NewHTTPFetcher creates a new HTTP fetcher.
func NewHTTPFetcher(url string, timeout time.Duration) *HTTPFetcher {
	return &HTTPFetcher{
		url:     url,
		timeout: timeout,
	}
}

func (f *HTTPFetcher) Fetch(ctx context.Context) ([]byte, string, error) {
	client := &http.Client{
		Timeout: f.timeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	return body, f.url, nil
}

// CachedFetcher wraps a Fetcher with caching behavior.
type CachedFetcher struct {
	fetcher Fetcher
	cache   *Cache
	source  string
}

// NewCachedFetcher creates a new cached fetcher that wraps the provided fetcher.
func NewCachedFetcher(f Fetcher, cache *Cache, source string) *CachedFetcher {
	return &CachedFetcher{
		fetcher: f,
		cache:   cache,
		source:  source,
	}
}

// Fetch retrieves data, checking cache first and storing results in cache.
// If refresh is true, bypasses cache and fetches fresh data.
func (c *CachedFetcher) Fetch(ctx context.Context, refresh bool) ([]byte, string, error) {
	if !refresh {
		if cachedData, cachedSource, found := c.cache.Get(c.source); found {
			return cachedData, cachedSource, nil
		}
	}

	data, sourceID, err := c.fetcher.Fetch(ctx)
	if err != nil {
		return nil, "", err
	}

	c.cache.Set(c.source, data, sourceID)
	return data, sourceID, nil
}
