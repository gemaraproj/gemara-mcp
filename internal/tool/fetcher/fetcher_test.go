// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFetcher is a test fetcher that returns predefined data.
type mockFetcher struct {
	data      []byte
	source    string
	err       error
	callCount int
}

func (m *mockFetcher) Fetch(ctx context.Context) ([]byte, string, error) {
	m.callCount++
	if m.err != nil {
		return nil, "", m.err
	}
	return m.data, m.source, nil
}

func TestCachedFetcher(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() *mockFetcher
		refresh   bool
		wantErr   bool
		validate  func(t *testing.T, data []byte, source string, callCount int)
	}{
		{
			name: "cache miss fetches from underlying fetcher",
			setupMock: func() *mockFetcher {
				return &mockFetcher{
					data:   []byte("test data"),
					source: "mock://test",
				}
			},
			refresh: false,
			wantErr: false,
			validate: func(t *testing.T, data []byte, source string, callCount int) {
				assert.Equal(t, []byte("test data"), data)
				assert.Equal(t, "mock://test", source)
				assert.Equal(t, 1, callCount, "should call underlying fetcher once")
			},
		},
		{
			name: "cache hit returns cached data",
			setupMock: func() *mockFetcher {
				return &mockFetcher{
					data:   []byte("cached data"),
					source: "mock://test",
				}
			},
			refresh: false,
			wantErr: false,
			validate: func(t *testing.T, data []byte, source string, callCount int) {
				assert.Equal(t, []byte("cached data"), data)
				assert.Equal(t, 1, callCount, "should only call underlying fetcher once (first call)")
			},
		},
		{
			name: "refresh bypasses cache",
			setupMock: func() *mockFetcher {
				return &mockFetcher{
					data:   []byte("fresh data"),
					source: "mock://test",
				}
			},
			refresh: true,
			wantErr: false,
			validate: func(t *testing.T, data []byte, source string, callCount int) {
				assert.Equal(t, []byte("fresh data"), data)
				assert.Equal(t, 2, callCount, "should call underlying fetcher twice (refresh bypasses cache)")
			},
		},
		{
			name: "fetch error propagates",
			setupMock: func() *mockFetcher {
				return &mockFetcher{
					err: errors.New("fetch failed"),
				}
			},
			refresh: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cache := NewCache(24 * time.Hour) // Test-specific TTL
			source := "test://source"

			mock := tt.setupMock()
			cf := NewCachedFetcher(mock, cache, source)

			// For cache hit test, make two calls
			if tt.name == "cache hit returns cached data" {
				// First call - should fetch
				data1, source1, err1 := cf.Fetch(ctx, false)
				require.NoError(t, err1, "first call should not error")
				assert.Equal(t, []byte("cached data"), data1)
				assert.Equal(t, "mock://test", source1)

				// Second call - should use cache
				data2, source2, err2 := cf.Fetch(ctx, false)
				require.NoError(t, err2, "second call should not error")
				assert.Equal(t, data1, data2, "cached data should match")
				assert.Equal(t, source1, source2, "cached source should match")
				assert.Equal(t, 1, mock.callCount, "should only call underlying fetcher once")
				return
			}

			// For refresh test
			if tt.name == "refresh bypasses cache" {
				// First call
				_, _, err1 := cf.Fetch(ctx, false)
				require.NoError(t, err1, "first call should not error")

				// Second call with refresh
				data2, source2, err2 := cf.Fetch(ctx, true)
				require.NoError(t, err2, "refresh call should not error")
				assert.Equal(t, []byte("fresh data"), data2)
				assert.Equal(t, "mock://test", source2)
				assert.Equal(t, 2, mock.callCount, "should call underlying fetcher twice")
				return
			}

			// Regular test execution
			data, sourceID, err := cf.Fetch(ctx, tt.refresh)

			if tt.wantErr {
				assert.Error(t, err, "should return error")
				return
			}

			require.NoError(t, err, "should not return error")
			if tt.validate != nil {
				tt.validate(t, data, sourceID, mock.callCount)
			}
		})
	}
}
