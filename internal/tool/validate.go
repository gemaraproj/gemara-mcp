// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"cuelang.org/go/cue"
	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
	"github.com/gemaraproj/gemara-mcp/internal/tool/schema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MetadataValidateGemaraArtifact describes the ValidateGemaraArtifact tool.
var MetadataValidateGemaraArtifact = &mcp.Tool{
	Name:        "validate_gemara_artifact",
	Description: "Validate a Gemara artifact YAML content against the Gemara CUE schema using the CUE registry module.",
	InputSchema: map[string]interface{}{
		"type":     "object",
		"required": []string{"artifact_content", "definition"},
		"properties": map[string]interface{}{
			"artifact_content": map[string]interface{}{
				"type":        "string",
				"description": "YAML content of the Gemara artifact to validate",
			},
			"definition": map[string]interface{}{
				"type":        "string",
				"description": "CUE definition name to validate against (e.g., '#ControlCatalog', '#GuidanceCatalog', '#Policy', '#EvaluationLog')",
			},
			"version": map[string]interface{}{
				"type":        "string",
				"description": "Version of the Gemara module to validate against (default: 'latest')",
			},
		},
	},
}

// InputValidateGemaraArtifact is the input for the ValidateGemaraArtifact tool.
type InputValidateGemaraArtifact struct {
	ArtifactContent string `json:"artifact_content"`
	Definition      string `json:"definition"`
	Version         string `json:"version"`
}

// OutputValidateGemaraArtifact is the output for the ValidateGemaraArtifact tool.
type OutputValidateGemaraArtifact struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message"`
}

// ValidateGemaraArtifact validates a Gemara artifact using the CUE Go SDK with the registry module.
// The returned *mcp.CallToolResult is always nil; the go-sdk derives the
// tool response from the OutputValidateGemaraArtifact struct.
func ValidateGemaraArtifact(ctx context.Context, _ *mcp.CallToolRequest, input InputValidateGemaraArtifact, cf *fetcher.CachedFetcher[cue.Value]) (*mcp.CallToolResult, OutputValidateGemaraArtifact, error) {
	// Validate inputs
	if input.ArtifactContent == "" {
		return nil, OutputValidateGemaraArtifact{}, fmt.Errorf("artifact_content is required")
	}
	if input.Definition == "" {
		return nil, OutputValidateGemaraArtifact{}, fmt.Errorf("definition is required")
	}

	// Ensure definition starts with #
	definition := input.Definition
	if !strings.HasPrefix(definition, "#") {
		definition = "#" + definition
	}

	slog.Info("validating artifact", "definition", definition, "content_length", len(input.ArtifactContent))

	cueVal, _, err := cf.Fetch(ctx, false)
	if err != nil {
		return nil, OutputValidateGemaraArtifact{}, fmt.Errorf("loading schema: %w", err)
	}

	result, err := schema.Validate(cueVal, definition, input.ArtifactContent)
	if err != nil {
		return nil, OutputValidateGemaraArtifact{}, err
	}

	slog.Info("validation complete", "definition", definition, "valid", result.Valid, "error_count", len(result.Errors))
	return nil, OutputValidateGemaraArtifact{
		Valid:   result.Valid,
		Errors:  result.Errors,
		Message: result.Message,
	}, nil
}
