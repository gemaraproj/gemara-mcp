// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
)

// mockFetcher is a test fetcher that returns predefined data.
type mockFetcher struct {
	data   []byte
	source string
	err    error
}

func (m *mockFetcher) Fetch(ctx context.Context) ([]byte, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	return m.data, m.source, nil
}

func TestGetLexicon(t *testing.T) {
	tests := []struct {
		name           string
		mockFetcher    *mockFetcher
		input          InputGetLexicon
		wantErr        bool
		wantEntryCount int
		validateOutput func(t *testing.T, output OutputGetLexicon)
	}{
		{
			name: "successful fetch and parse",
			mockFetcher: &mockFetcher{
				data: []byte(`- term: Assessment
  definition: Atomic process used to determine a resource's compliance
  references: ["Layer 5"]
- term: Control
  definition: Safeguard or countermeasure
  references: ["Layer 2"]`),
				source: "mock://lexicon.yaml",
			},
			input:          InputGetLexicon{Refresh: false},
			wantErr:        false,
			wantEntryCount: 2,
			validateOutput: func(t *testing.T, output OutputGetLexicon) {
				assert.Len(t, output.Entries, 2, "should have 2 entries")
				assert.Equal(t, "Assessment", output.Entries[0].Term, "first term should be Assessment")
				assert.Equal(t, "Control", output.Entries[1].Term, "second term should be Control")
			},
		},
		{
			name: "fetch error returns error",
			mockFetcher: &mockFetcher{
				err: errors.New("fetch failed"),
			},
			input:   InputGetLexicon{Refresh: false},
			wantErr: true,
		},
		{
			name: "invalid YAML returns error",
			mockFetcher: &mockFetcher{
				data:   []byte("invalid: yaml: content: [unclosed"),
				source: "mock://invalid.yaml",
			},
			input:   InputGetLexicon{Refresh: false},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cache := fetcher.NewCache(24 * time.Hour) // Test-specific TTL
			cf := fetcher.NewCachedFetcher(tt.mockFetcher, cache, "mock://source")

			_, output, err := GetLexicon(ctx, nil, tt.input, cf)

			if tt.wantErr {
				assert.Error(t, err, "should return error")
				return
			}

			require.NoError(t, err, "should not return error")
			assert.Len(t, output.Entries, tt.wantEntryCount, "entry count should match")
			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}
