// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/yaml"
	"cuelang.org/go/mod/modconfig"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	gemaraModulePath = "github.com/gemaraproj/gemara@latest"
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
				"description": "CUE definition name to validate against (e.g., '#ControlCatalog', '#GuidanceDocument', '#Policy', '#EvaluationLog')",
			},
		},
	},
}

// InputValidateGemaraArtifact is the input for the ValidateGemaraArtifact tool.
type InputValidateGemaraArtifact struct {
	ArtifactContent string `json:"artifact_content"`
	Definition      string `json:"definition"`
}

// OutputValidateGemaraArtifact is the output for the ValidateGemaraArtifact tool.
type OutputValidateGemaraArtifact struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message"`
}

// ValidateGemaraArtifact validates a Gemara artifact using the CUE Go SDK with the registry module.
func ValidateGemaraArtifact(ctx context.Context, _ *mcp.CallToolRequest, input InputValidateGemaraArtifact) (*mcp.CallToolResult, OutputValidateGemaraArtifact, error) {
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

	reg, err := modconfig.NewRegistry(nil)
	if err != nil {
		return nil, OutputValidateGemaraArtifact{}, fmt.Errorf("failed to create CUE registry: %w", err)
	}

	// Pass the module path as an argument to load it from the registry
	buildInstances := load.Instances([]string{gemaraModulePath}, &load.Config{
		Registry: reg,
	})

	if len(buildInstances) == 0 {
		return nil, OutputValidateGemaraArtifact{}, fmt.Errorf("failed to load module: no instances returned")
	}

	if err := buildInstances[0].Err; err != nil {
		return nil, OutputValidateGemaraArtifact{}, fmt.Errorf("failed to load module: %w", err)
	}

	cueCtx := cuecontext.New()
	schema := cueCtx.BuildInstance(buildInstances[0])
	if err := schema.Err(); err != nil {
		return nil, OutputValidateGemaraArtifact{}, fmt.Errorf("failed to build schema: %w", err)
	}

	entrypointPath := cue.ParsePath(definition)
	entrypoint := schema.LookupPath(entrypointPath)
	if !entrypoint.Exists() {
		return nil, OutputValidateGemaraArtifact{}, fmt.Errorf("definition %s not found in schema", definition)
	}

	yamlFile, err := yaml.Extract("artifact.yaml", input.ArtifactContent)
	if err != nil {
		// Invalid YAML should result in validation failure, not a function error
		output := OutputValidateGemaraArtifact{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("Failed to parse YAML: %v", err)},
			Message: fmt.Sprintf("Validation failed: invalid YAML: %v", err),
		}
		return nil, output, nil
	}

	data := cueCtx.BuildFile(yamlFile)
	if err := data.Err(); err != nil {
		// Data build errors should result in validation failure
		output := OutputValidateGemaraArtifact{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("Failed to build data instance: %v", err)},
			Message: fmt.Sprintf("Validation failed: %v", err),
		}
		return nil, output, nil
	}

	unified := entrypoint.Unify(data)

	if err := unified.Validate(cue.Concrete(true)); err != nil {
		errorOutput := err.Error()
		errorLines := strings.Split(strings.TrimSpace(errorOutput), "\n")

		// Filter out empty lines
		var errors []string
		for _, line := range errorLines {
			if strings.TrimSpace(line) != "" {
				errors = append(errors, line)
			}
		}

		output := OutputValidateGemaraArtifact{
			Valid:   false,
			Errors:  errors,
			Message: fmt.Sprintf("Validation failed: %v", err),
		}
		return nil, output, nil
	}

	output := OutputValidateGemaraArtifact{
		Valid:   true,
		Errors:  []string{},
		Message: "Artifact is valid",
	}

	return nil, output, nil
}
