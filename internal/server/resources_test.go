// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"testing"
	"time"

	"github.com/gemaraproj/gemara-mcp/internal/server/fetcher"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubByteFetcher returns a scripted sequence of responses, one per Fetch call.
type stubByteFetcher struct {
	responses [][]byte
	calls     int
}

func (s *stubByteFetcher) Fetch(context.Context) ([]byte, string, error) {
	r := s.responses[s.calls]
	s.calls++
	return r, "stub", nil
}

func setupAdvisorySession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	mode, err := NewAdvisoryMode(1 * time.Hour)
	require.NoError(t, err)
	server := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.0"},
		&mcp.ServerOptions{Instructions: mode.Description()},
	)
	mode.Register(server)
	return connectSession(t, server)
}

func TestReadSchemaDocsTemplateResourceInvalidVersion(t *testing.T) {
	session := setupAdvisorySession(t)
	_, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "gemara://schema/definitions?version=not-semver",
	})
	require.Error(t, err)
}

func TestParseLexicon(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:  "valid lexicon",
			input: EmbeddedLexicon,
		},
		{
			name:        "invalid YAML",
			input:       "not: [valid: yaml",
			wantErr:     true,
			errContains: "unmarshalling",
		},
		{
			name: "missing title",
			input: `metadata:
  id: test
  type: Lexicon
  gemara-version: "1.0.0"
  description: test
  author:
    id: test
    name: Test
    type: Human
terms:
  - id: t1
    title: Term
    definition: def`,
			wantErr:     true,
			errContains: "title",
		},
		{
			name: "empty terms",
			input: `title: Test Lexicon
metadata:
  id: test
  type: Lexicon
  gemara-version: "1.0.0"
  description: test
  author:
    id: test
    name: Test
    type: Human
terms: []`,
			wantErr:     true,
			errContains: "no terms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex, err := parseLexicon([]byte(tt.input))
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, lex.Title)
			assert.NotEmpty(t, lex.Terms)
		})
	}
}

func TestIsValidAbout(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "embedded content", input: EmbeddedAbout, want: true},
		{name: "leading whitespace", input: "\n\n# Gemara\n\nintro", want: true},
		{name: "leading utf8 bom", input: "\ufeff# Gemara\n\nintro", want: true},
		{name: "heading variation", input: "# Gemara: Overview\n\nintro", want: true},
		{name: "empty", input: "", want: false},
		{name: "whitespace only", input: "   \n\t ", want: false},
		{name: "html error page", input: "<!DOCTYPE html><html>404</html>", want: false},
		{name: "wrong heading", input: "# Something Else", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidAbout([]byte(tt.input)))
		})
	}
}

func TestEmbeddedAboutIsValid(t *testing.T) {
	assert.True(t, isValidAbout([]byte(EmbeddedAbout)))
	assert.Contains(t, EmbeddedAbout, "Seven-Layer Model")
}

// TestValidatingFetcherDoesNotCacheInvalidContent guards against cache
// poisoning: an invalid 200 response must surface as an error before the cache
// stores it, so a later recovered upstream is served instead of a pinned bad
// body.
func TestValidatingFetcherDoesNotCacheInvalidContent(t *testing.T) {
	stub := &stubByteFetcher{responses: [][]byte{
		[]byte("<!DOCTYPE html><html>502</html>"),
		[]byte("# Gemara\n\nvalid content"),
	}}
	vf := validatingFetcher{inner: stub, validate: func(data []byte) error {
		if !isValidAbout(data) {
			return assert.AnError
		}
		return nil
	}}
	cf := fetcher.NewCachedFetcher[[]byte](vf, fetcher.NewCache[[]byte](1*time.Hour), "about")

	_, _, err := cf.Fetch(context.Background(), false)
	require.Error(t, err, "invalid content should surface as an error")

	data, _, err := cf.Fetch(context.Background(), false)
	require.NoError(t, err, "recovered upstream must not be blocked by a poisoned cache")
	assert.Equal(t, "# Gemara\n\nvalid content", string(data))
}

func TestEmbeddedLexiconConformsToSDKType(t *testing.T) {
	lex, err := parseLexicon([]byte(EmbeddedLexicon))
	require.NoError(t, err)

	assert.Equal(t, "Gemara Lexicon", lex.Title)
	assert.Equal(t, "gemara-lexicon", lex.Metadata.Id)
	assert.Greater(t, len(lex.Terms), 0)

	for _, term := range lex.Terms {
		assert.NotEmpty(t, term.Id, "term missing id")
		assert.NotEmpty(t, term.Title, "term %s missing title", term.Id)
		assert.NotEmpty(t, term.Definition, "term %s missing definition", term.Id)
	}
}
