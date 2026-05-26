// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
