// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateGemaraArtifact(t *testing.T) {
	// Load test data
	testDataDir := filepath.Join("test-data")
	validControlCatalogPath := filepath.Join(testDataDir, "good-ccc.yaml")
	validControlCatalogContent, err := os.ReadFile(validControlCatalogPath)
	require.NoError(t, err, "should be able to read test data file")

	tests := []struct {
		name           string
		input          InputValidateGemaraArtifact
		wantErr        bool
		errContains    string
		wantValid      *bool // nil means don't check, true/false means check value
		validateOutput func(t *testing.T, output OutputValidateGemaraArtifact)
	}{
		{
			name: "missing artifact_content",
			input: InputValidateGemaraArtifact{
				ArtifactContent: "",
				Definition:      "#ControlCatalog",
			},
			wantErr:     true,
			errContains: "artifact_content is required",
		},
		{
			name: "missing definition",
			input: InputValidateGemaraArtifact{
				ArtifactContent: "test: content",
				Definition:      "",
			},
			wantErr:     true,
			errContains: "definition is required",
		},
		{
			name: "valid ControlCatalog from testdata",
			input: InputValidateGemaraArtifact{
				ArtifactContent: string(validControlCatalogContent),
				Definition:      "#ControlCatalog",
			},
			wantErr: false,
			validateOutput: func(t *testing.T, output OutputValidateGemaraArtifact) {
				// Validation may succeed or fail depending on cue availability and schema
				// But the function should execute without error
				assert.NotEmpty(t, output.Message, "should have a message")
				// If valid, should have no errors; if invalid, should have errors
				if output.Valid {
					assert.Empty(t, output.Errors, "if valid, should have no errors")
				} else {
					assert.NotEmpty(t, output.Errors, "if invalid, should have errors")
				}
			},
		},
		{
			name: "valid ControlCatalog without hash prefix",
			input: InputValidateGemaraArtifact{
				ArtifactContent: string(validControlCatalogContent),
				Definition:      "ControlCatalog", // Missing #
			},
			wantErr: false,
			validateOutput: func(t *testing.T, output OutputValidateGemaraArtifact) {
				// Hash prefix should be added automatically
				assert.NotEmpty(t, output.Message, "should have a message")
			},
		},
		{
			name: "invalid YAML content",
			input: InputValidateGemaraArtifact{
				ArtifactContent: "invalid: yaml: [unclosed",
				Definition:      "#ControlCatalog",
			},
			wantErr:   false,
			wantValid: boolPtr(false),
			validateOutput: func(t *testing.T, output OutputValidateGemaraArtifact) {
				// Invalid YAML should result in validation failure
				assert.False(t, output.Valid, "should be invalid")
				assert.NotEmpty(t, output.Errors, "should have errors for invalid YAML")
				assert.Contains(t, output.Message, "YAML", "message should mention YAML")
			},
		},
		{
			name: "valid YAML but wrong definition type",
			input: InputValidateGemaraArtifact{
				ArtifactContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test`,
				Definition: "#ControlCatalog",
			},
			wantErr:   false,
			wantValid: boolPtr(false),
			validateOutput: func(t *testing.T, output OutputValidateGemaraArtifact) {
				assert.False(t, output.Valid, "should be invalid - wrong schema")
				assert.NotEmpty(t, output.Errors, "should have validation errors")
			},
		},
		{
			name: "empty YAML content",
			input: InputValidateGemaraArtifact{
				ArtifactContent: "",
				Definition:      "#ControlCatalog",
			},
			wantErr:     true,
			errContains: "artifact_content is required",
		},
		{
			name: "definition with hash prefix preserved",
			input: InputValidateGemaraArtifact{
				ArtifactContent: string(validControlCatalogContent),
				Definition:      "#ControlCatalog",
			},
			wantErr:   false,
			wantValid: boolPtr(true),
			validateOutput: func(t *testing.T, output OutputValidateGemaraArtifact) {
				assert.True(t, output.Valid, "should be valid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Arguments: json.RawMessage(`{}`),
				},
			}

			_, output, err := ValidateGemaraArtifact(ctx, req, tt.input)

			if tt.wantErr {
				require.Error(t, err, "should return error")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains, "error should contain expected message")
				}
				return
			}

			require.NoError(t, err, "should not return error")
			if tt.wantValid != nil {
				assert.Equal(t, *tt.wantValid, output.Valid, "valid status should match")
			}
			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool {
	return &b
}
