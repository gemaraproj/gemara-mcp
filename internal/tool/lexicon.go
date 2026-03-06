// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
	"github.com/goccy/go-yaml"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LexiconEntry represents a single term in the Gemara Lexicon.
type LexiconEntry struct {
	Term       string   `json:"term" yaml:"term"`
	Definition string   `json:"definition" yaml:"definition"`
	References []string `json:"references" yaml:"references"`
}

// OutputGetLexicon is the output for the GetLexicon tool.
type OutputGetLexicon struct {
	Entries []LexiconEntry `json:"entries"`
	Source  string         `json:"source"`
}

// MetadataGetLexicon describes the GetLexicon tool.
var MetadataGetLexicon = &mcp.Tool{
	Name:        "get_lexicon",
	Description: "Retrieve the Gemara Lexicon containing definitions of terms used in the Gemara model.",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"refresh": map[string]interface{}{
				"type":        "boolean",
				"description": "Force refresh of lexicon cache (default: false)",
			},
			"version": map[string]interface{}{
				"type":        "string",
				"description": "Version of the Gemara lexicon to retrieve (e.g., 'v0.19.1')",
			},
		},
	},
}

// InputGetLexicon is the input for the GetLexicon tool.
type InputGetLexicon struct {
	Refresh bool   `json:"refresh"`
	Version string `json:"version"`
}

// GetLexicon retrieves the Gemara Lexicon using the specified cached fetcher.
func GetLexicon(ctx context.Context, _ *mcp.CallToolRequest, input InputGetLexicon, cachedFetcher *fetcher.CachedFetcher[[]byte]) (*mcp.CallToolResult, OutputGetLexicon, error) {
	slog.Info("fetching lexicon", "refresh", input.Refresh)

	data, sourceID, err := cachedFetcher.Fetch(ctx, input.Refresh)
	if err != nil {
		return nil, OutputGetLexicon{}, fmt.Errorf("failed to fetch lexicon: %w", err)
	}

	var entries []LexiconEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, OutputGetLexicon{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	slog.Info("lexicon loaded", "entry_count", len(entries), "source", sourceID)
	return nil, OutputGetLexicon{
		Entries: entries,
		Source:  sourceID,
	}, nil
}
