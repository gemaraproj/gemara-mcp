// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"

	"github.com/gemaraproj/gemara-mcp/internal/evidence"
	"github.com/gemaraproj/gemara-mcp/internal/evidence/parsers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MetadataParseGovernanceDocument describes the parse_governance_document tool.
var MetadataParseGovernanceDocument = &mcp.Tool{
	Name: "parse_governance_document",
	Description: "Parse a governance or technical configuration document and return schema-aligned " +
		"candidates for Gemara artifact generation. " +
		"Supported formats: markdown, yaml, json, kubernetes, dockerfile. " +
		"Each candidate includes a target schema field, a proposed value, its source reference, " +
		"and a confidence score. High-confidence candidates (≥0.7) are suitable for Tier 1 " +
		"(automated) artifact generation. Lower-confidence candidates should be reviewed by a human " +
		"(Tier 2) before inclusion.",
	InputSchema: map[string]interface{}{
		"type":     "object",
		"required": []string{"content"},
		"properties": map[string]interface{}{
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Raw content of the document to parse",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Format hint for the document. One of: markdown, yaml, json, kubernetes. If omitted, auto-detection is used.",
				"enum":        []string{"markdown", "yaml", "json", "kubernetes"},
			},
			"source_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional identifier for the document (file path, URL, etc.) used in candidate source references.",
			},
		},
	},
}

// InputParseGovernanceDocument is the input for the ParseGovernanceDocument tool.
type InputParseGovernanceDocument struct {
	Content  string `json:"content"`
	Format   string `json:"format"`
	SourceID string `json:"source_id"`
}

// OutputParseGovernanceDocument is the output for the ParseGovernanceDocument tool.
type OutputParseGovernanceDocument struct {
	Candidates  []evidence.SchemaCandidate `json:"candidates"`
	ParserUsed  string                     `json:"parser_used"`
	TotalChunks int                        `json:"total_chunks"`
}

// defaultPipeline builds a Pipeline with all default parsers registered..
func defaultPipeline() *evidence.Pipeline {
	return evidence.NewPipeline(
		parsers.NewKubernetesParser(),
		parsers.NewMarkdownParser(),
		parsers.NewYAMLParser(),
	)
}

// ParseGovernanceDocument runs the evidence pipeline over the provided document
// and returns schema-aligned candidates for artifact generation.
func ParseGovernanceDocument(ctx context.Context, _ *mcp.CallToolRequest, input InputParseGovernanceDocument) (*mcp.CallToolResult, OutputParseGovernanceDocument, error) {
	if input.Content == "" {
		return nil, OutputParseGovernanceDocument{}, fmt.Errorf("content is required")
	}

	sourceID := input.SourceID
	if sourceID == "" {
		sourceID = "unknown"
	}

	src := evidence.EvidenceSource{
		Content: []byte(input.Content),
		Format:  input.Format,
		ID:      sourceID,
	}

	pipeline := defaultPipeline()
	result, err := pipeline.RunWithMeta(ctx, src)
	if err != nil {
		return nil, OutputParseGovernanceDocument{}, err
	}

	return nil, OutputParseGovernanceDocument{
		Candidates:  result.Candidates,
		ParserUsed:  result.ParserUsed,
		TotalChunks: result.ChunkCount,
	}, nil
}
