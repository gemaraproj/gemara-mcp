// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// OutputGetSchemaDocs is the output for the GetSchemaDocs tool.
type OutputGetSchemaDocs struct {
	Documentation string `json:"documentation"`
	URL           string `json:"url"`
}

// MetadataGetSchemaDocs describes the GetSchemaDocs tool.
var MetadataGetSchemaDocs = &mcp.Tool{
	Name:        "get_schema_docs",
	Description: "Retrieve schema documentation for the Gemara CUE module from the CUE registry.",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"refresh": map[string]interface{}{
				"type":        "boolean",
				"description": "Force refresh of schema docs cache (default: false)",
			},
			"version": map[string]interface{}{
				"type":        "string",
				"description": "Version of the Gemara module (default: 'latest')",
			},
		},
	},
}

// InputGetSchemaDocs is the input for the GetSchemaDocs tool.
type InputGetSchemaDocs struct {
	Refresh bool   `json:"refresh"`
	Version string `json:"version"`
}

// GetSchemaDocs retrieves schema documentation using the specified cached fetcher.
func GetSchemaDocs(ctx context.Context, _ *mcp.CallToolRequest, input InputGetSchemaDocs, cachedFetcher *fetcher.CachedFetcher) (*mcp.CallToolResult, OutputGetSchemaDocs, error) {
	data, sourceID, err := cachedFetcher.Fetch(ctx, input.Refresh)
	if err != nil {
		return nil, OutputGetSchemaDocs{}, err
	}

	output := OutputGetSchemaDocs{
		Documentation: string(data),
		URL:           sourceID,
	}

	return nil, output, nil
}
