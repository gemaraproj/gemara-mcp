// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	LexiconResourceURI      = "https://gemara.openssf.org/model/02-definitions"
	LexiconResourceURIAlias = "gemara://lexicon"
	lexiconResourceURI      = LexiconResourceURI
)

// MetadataLexiconResource describes the Lexicon resource with the canonical URL.
var MetadataLexiconResource = &mcp.Resource{
	Name:        "lexicon",
	URI:         LexiconResourceURI,
	Title:       "Gemara Lexicon",
	Description: "The Gemara Lexicon containing definitions of terms used in the Gemara framework.",
	MIMEType:    "application/json",
}

// MetadataLexiconResourceAlias describes the Lexicon resource with the alias URI.
var MetadataLexiconResourceAlias = &mcp.Resource{
	Name:        "lexicon",
	URI:         LexiconResourceURIAlias,
	Title:       "Gemara Lexicon",
	Description: "The Gemara Lexicon containing definitions of terms used in the Gemara framework.",
	MIMEType:    "application/json",
}

// HandleLexiconResource reads the cached Lexicon resource.
func HandleLexiconResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Ensure lexicon is loaded by fetching if cache is empty or expired
	if len(lexiconCache) == 0 || lexiconCacheTime.IsZero() || time.Since(lexiconCacheTime) >= lexiconCacheTTL {
		entries, err := fetchLexiconFromURL(ctx, lexiconURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch lexicon: %w", err)
		}

		// Update cache
		lexiconCache = entries
		lexiconCacheTime = time.Now()
	}

	// Marshal lexicon to JSON
	lexiconJSON, err := json.Marshal(lexiconCache)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal lexicon: %w", err)
	}

	// Use the requested URI in the response
	requestedURI := req.Params.URI
	if requestedURI == "" {
		requestedURI = LexiconResourceURI
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      requestedURI,
				MIMEType: "application/json",
				Text:     string(lexiconJSON),
			},
		},
	}, nil
}
