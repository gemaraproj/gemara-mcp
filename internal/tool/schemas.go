// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"log/slog"

	"cuelang.org/go/cue"
	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
	"github.com/gemaraproj/gemara-mcp/internal/tool/schema"
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
	Description: "Retrieve schema definitions for the Gemara CUE module using the CUE registry.",
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

// GetSchemaDocs loads the Gemara CUE module and returns formatted schema definitions.
func GetSchemaDocs(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input InputGetSchemaDocs,
	cf *fetcher.CachedFetcher[cue.Value],
) (*mcp.CallToolResult, OutputGetSchemaDocs, error) {
	val, source, err := cf.Fetch(ctx, input.Refresh)
	if err != nil {
		return nil, OutputGetSchemaDocs{}, err
	}

	defs, err := schema.FormatDefinitions(val)
	if err != nil {
		return nil, OutputGetSchemaDocs{}, err
	}

	slog.Info("schema docs loaded", "source", source, "length", len(defs))
	return nil, OutputGetSchemaDocs{
		Documentation: defs,
		URL:           source,
	}, nil
}
