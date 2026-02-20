// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// defaultMaxResponseBytes limits response body reads.
var defaultMaxResponseBytes int64 = 4 * 1024 * 1024 // 4 MiB

// Fetcher is a generic interface for fetching raw data from a source.
type Fetcher interface {
	// Fetch retrieves raw data from the source and returns the data and source identifier.
	Fetch(ctx context.Context) ([]byte, string, error)
}

// HTTPFetcher fetches data from an HTTP URL.
type HTTPFetcher struct {
	fetchURL string

	// MaxResponseBytes limits the number of response body bytes read.
	// If zero or negative, defaultMaxResponseBytes is used.
	MaxResponseBytes int64
}

// NewHTTPFetcher creates an HTTP fetcher by building a URL from the
// trusted URLBuilder and the given version.
func NewHTTPFetcher(builder *URLBuilder, version string) (*HTTPFetcher, error) {
	fetchURL, err := builder.Build(version)
	if err != nil {
		return nil, fmt.Errorf("building fetch URL: %w", err)
	}
	return &HTTPFetcher{fetchURL: fetchURL}, nil
}

// URL returns the resolved fetch URL.
func (f *HTTPFetcher) URL() string {
	return f.fetchURL
}

func (f *HTTPFetcher) Fetch(ctx context.Context) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.fetchURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf(
			"%s %q: response status code %d: %s",
			resp.Request.Method, resp.Request.URL,
			resp.StatusCode, http.StatusText(resp.StatusCode),
		)
	}

	body, err := io.ReadAll(limitReader(resp.Body, f.MaxResponseBytes))
	if err != nil {
		return nil, "", fmt.Errorf("reading response body: %w", err)
	}

	return body, f.fetchURL, nil
}

// limitReader returns a Reader that stops after n bytes.
// If n is zero or negative, defaultMaxResponseBytes is used.
func limitReader(r io.Reader, n int64) io.Reader {
	if n <= 0 {
		n = defaultMaxResponseBytes
	}
	return io.LimitReader(r, n)
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
