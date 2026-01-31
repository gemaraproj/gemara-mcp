// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLexicon(t *testing.T) {
	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		input          InputGetLexicon
		wantErr        bool
		wantCached     bool
		wantEntryCount int
		validateOutput func(t *testing.T, output OutputGetLexicon)
	}{
		{
			name: "successful fetch",
			setupServer: func() *httptest.Server {
				mockYAML := `- term: Assessment
  definition: Atomic process used to determine a resource's compliance
  references: ["Layer 5"]
- term: Control
  definition: Safeguard or countermeasure
  references: ["Layer 2"]`
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/yaml")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(mockYAML))
				}))
			},
			input:          InputGetLexicon{Refresh: false},
			wantErr:        false,
			wantCached:     false,
			wantEntryCount: 2,
			validateOutput: func(t *testing.T, output OutputGetLexicon) {
				assert.Len(t, output.Entries, 2, "should have 2 entries")
				assert.Equal(t, "Assessment", output.Entries[0].Term, "first term should be Assessment")
			},
		},
		{
			name: "cache hit on second call",
			setupServer: func() *httptest.Server {
				mockYAML := `- term: Test
  definition: A test term
  references: []`
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(mockYAML))
				}))
			},
			input:          InputGetLexicon{Refresh: false},
			wantErr:        false,
			wantCached:     true, // Second call should be cached
			wantEntryCount: 1,
			validateOutput: nil, // Will be set in test
		},
		{
			name: "cache refresh bypasses cache",
			setupServer: func() *httptest.Server {
				mockYAML := `- term: Test
  definition: A test term
  references: []`
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(mockYAML))
				}))
			},
			input:          InputGetLexicon{Refresh: true},
			wantErr:        false,
			wantCached:     false,
			wantEntryCount: 1,
			validateOutput: nil,
		},
		{
			name: "HTTP error returns error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			input:   InputGetLexicon{Refresh: false},
			wantErr: true,
		},
		{
			name: "invalid YAML returns error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("invalid: yaml: content: [unclosed"))
				}))
			},
			input:   InputGetLexicon{Refresh: false},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset cache for each test
			lexiconCache = nil
			lexiconCacheTime = time.Time{}

			server := tt.setupServer()
			defer server.Close()

			ctx := context.Background()

			// For cache hit test, make two calls
			if tt.name == "cache hit on second call" {
				// First call - should fetch
				_, output1, err1 := getLexiconWithURL(ctx, InputGetLexicon{Refresh: false}, server.URL)
				require.NoError(t, err1, "first call should not error")
				assert.False(t, output1.Cached, "first call should not be cached")

				// Second call - should use cache
				_, output2, err2 := getLexiconWithURL(ctx, InputGetLexicon{Refresh: false}, server.URL)
				require.NoError(t, err2, "second call should not error")
				assert.True(t, output2.Cached, "second call should be cached")
				assert.Equal(t, len(output1.Entries), len(output2.Entries), "cached entries should match")
				return
			}

			// For cache refresh test, make two calls with refresh=true on second
			if tt.name == "cache refresh bypasses cache" {
				// First call
				_, _, err1 := getLexiconWithURL(ctx, InputGetLexicon{Refresh: false}, server.URL)
				require.NoError(t, err1, "first call should not error")

				// Second call with refresh
				_, output2, err2 := getLexiconWithURL(ctx, InputGetLexicon{Refresh: true}, server.URL)
				require.NoError(t, err2, "refresh call should not error")
				assert.False(t, output2.Cached, "refresh call should not be cached")
				return
			}

			// Regular test execution - use getLexiconWithURL to pass test server URL
			_, output, err := getLexiconWithURL(ctx, tt.input, server.URL)

			if tt.wantErr {
				assert.Error(t, err, "should return error")
				return
			}

			require.NoError(t, err, "should not return error")
			assert.Equal(t, tt.wantCached, output.Cached, "cached status should match")
			assert.Len(t, output.Entries, tt.wantEntryCount, "entry count should match")
			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}
