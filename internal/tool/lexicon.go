// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	lexiconURL      = "https://raw.githubusercontent.com/gemaraproj/gemara/main/docs/lexicon.yaml"
	httpTimeout     = 30 * time.Second
	lexiconCacheTTL = 24 * time.Hour // Cache for 24 hours since lexicon changes infrequently
)

var (
	lexiconCache     []LexiconEntry
	lexiconCacheTime time.Time
)

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
		},
	},
}

// InputGetLexicon is the input for the GetLexicon tool.
type InputGetLexicon struct {
	Refresh bool `json:"refresh"`
}

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
	Cached  bool           `json:"cached"`
}

// GetLexicon retrieves the Gemara Lexicon using the resource handler.
func GetLexicon(ctx context.Context, _ *mcp.CallToolRequest, input InputGetLexicon) (*mcp.CallToolResult, OutputGetLexicon, error) {
	// If refresh is requested, fetch fresh data and update cache
	if input.Refresh {
		entries, err := fetchLexiconFromURL(ctx, lexiconURL)
		if err != nil {
			return nil, OutputGetLexicon{}, err
		}

		// Update cache
		lexiconCache = entries
		lexiconCacheTime = time.Now()

		output := OutputGetLexicon{
			Entries: entries,
			Source:  lexiconURL,
			Cached:  false,
		}
		return nil, output, nil
	}

	// Otherwise, use the resource handler which will use cached data or fetch if needed
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: lexiconResourceURI,
		},
	}

	result, err := HandleLexiconResource(ctx, req)
	if err != nil {
		return nil, OutputGetLexicon{}, fmt.Errorf("failed to read lexicon resource: %w", err)
	}

	if len(result.Contents) == 0 {
		return nil, OutputGetLexicon{}, fmt.Errorf("resource returned no contents")
	}

	var entries []LexiconEntry
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &entries); err != nil {
		return nil, OutputGetLexicon{}, fmt.Errorf("failed to parse lexicon JSON: %w", err)
	}

	// Determine if data was cached (check if it was already cached before resource call)
	wasCached := !lexiconCacheTime.IsZero() && time.Since(lexiconCacheTime) < lexiconCacheTTL

	output := OutputGetLexicon{
		Entries: entries,
		Source:  lexiconURL,
		Cached:  wasCached,
	}

	return nil, output, nil
}

// fetchLexiconFromURL fetches the lexicon from the given URL.
func fetchLexiconFromURL(ctx context.Context, url string) ([]LexiconEntry, error) {
	client := &http.Client{
		Timeout: httpTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lexicon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var entries []LexiconEntry
	if err := yaml.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return entries, nil
}

// getLexiconWithURL retrieves the lexicon from the specified URL (used for testing).
func getLexiconWithURL(ctx context.Context, input InputGetLexicon, url string) (*mcp.CallToolResult, OutputGetLexicon, error) {
	if !input.Refresh && !lexiconCacheTime.IsZero() && time.Since(lexiconCacheTime) < lexiconCacheTTL {
		output := OutputGetLexicon{
			Entries: lexiconCache,
			Source:  url,
			Cached:  true,
		}
		return nil, output, nil
	}

	entries, err := fetchLexiconFromURL(ctx, url)
	if err != nil {
		return nil, OutputGetLexicon{}, err
	}

	// Update cache
	lexiconCache = entries
	lexiconCacheTime = time.Now()

	output := OutputGetLexicon{
		Entries: entries,
		Source:  url,
		Cached:  false,
	}

	return nil, output, nil
}
