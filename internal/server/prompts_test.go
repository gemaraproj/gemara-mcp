// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testLexicon    = "test-lexicon-content"
	testSchemaDocs = "test-schema-docs-content"
)

func mockLexiconFetcher(_ context.Context) (string, error) {
	return testLexicon, nil
}

func failingLexiconFetcher(_ context.Context) (string, error) {
	return "", fmt.Errorf("lexicon fetch error")
}

func mockSchemaFetcher(_ context.Context) (string, error) {
	return testSchemaDocs, nil
}

func failingSchemaFetcher(_ context.Context) (string, error) {
	return "", fmt.Errorf("network error")
}

func assertEmbeddedResources(t *testing.T, messages []*mcp.PromptMessage) {
	t.Helper()
	require.GreaterOrEqual(t, len(messages), 2, "need at least 2 messages for embedded resources")

	lexiconMsg := messages[0]
	assert.Equal(t, mcp.Role("user"), lexiconMsg.Role)
	lexiconRes, ok := lexiconMsg.Content.(*mcp.EmbeddedResource)
	require.True(t, ok, "first message should be EmbeddedResource")
	assert.Equal(t, LexiconResourceURI, lexiconRes.Resource.URI)
	assert.Equal(t, "text/yaml", lexiconRes.Resource.MIMEType)
	assert.Equal(t, testLexicon, lexiconRes.Resource.Text)

	schemaMsg := messages[1]
	assert.Equal(t, mcp.Role("user"), schemaMsg.Role)
	schemaRes, ok := schemaMsg.Content.(*mcp.EmbeddedResource)
	require.True(t, ok, "second message should be EmbeddedResource")
	assert.Equal(t, SchemaDocsResourceURI, schemaRes.Resource.URI)
	assert.Equal(t, "text/plain", schemaRes.Resource.MIMEType)
	assert.Equal(t, testSchemaDocs, schemaRes.Resource.Text)
}

func TestNewThreatAssessmentHandler(t *testing.T) {
	handler := NewThreatAssessmentHandler(mockLexiconFetcher, mockSchemaFetcher)

	tests := []struct {
		name           string
		arguments      map[string]string
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, result *mcp.GetPromptResult)
	}{
		{
			name: "successful prompt generation with embedded resources",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "container runtime")
				assert.Contains(t, result.Description, "ACME.PLAT.CR")
				require.Len(t, result.Messages, 5, "2 embedded resources + 3 text messages")

				assertEmbeddedResources(t, result.Messages)

				instructionMsg := result.Messages[2]
				assert.Equal(t, mcp.Role("user"), instructionMsg.Role)
				text := instructionMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "container runtime")
				assert.Contains(t, text, "ACME.PLAT.CR")
				assert.Contains(t, text, "## Embedded Resources")
				assert.Contains(t, text, "## Available Tool")
				assert.Contains(t, text, "**Catalog Import**")
				assert.Contains(t, text, "**Scope and Metadata**")
				assert.Contains(t, text, "**Identify Capabilities**")
				assert.Contains(t, text, "**Identify Threats**")
				assert.Contains(t, text, "**Assemble and Validate**")
				assert.Contains(t, text, "**Next Steps**")
				assert.Contains(t, text, "validate_gemara_artifact")
				assert.Contains(t, text, "Privateer")

				assistantMsg := result.Messages[3]
				assert.Equal(t, mcp.Role("assistant"), assistantMsg.Role)
				assistantText := assistantMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, assistantText, "container runtime")
				assert.Contains(t, assistantText, "FINOS CCC Core")
				assert.Contains(t, assistantText, "Step 1: Catalog Import")
				assert.Contains(t, assistantText, "reply \"yes\"")

				userMsg := result.Messages[4]
				assert.Equal(t, mcp.Role("user"), userMsg.Role)
				userText := userMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, userText, "container runtime")
				assert.Contains(t, userText, "ACME.PLAT.CR")
			},
		},
		{
			name: "component with dots hyphens underscores accepted",
			arguments: map[string]string{
				"component": "kube-apiserver_v2.1",
				"id_prefix": "ACME.PLAT.KA",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "kube-apiserver_v2.1")
			},
		},
		{
			name: "missing component argument",
			arguments: map[string]string{
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "missing id_prefix argument",
			arguments: map[string]string{
				"component": "API gateway",
			},
			wantErr:     true,
			errContains: "id_prefix",
		},
		{
			name:        "both arguments missing",
			arguments:   map[string]string{},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "empty component string",
			arguments: map[string]string{
				"component": "",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "empty id_prefix string",
			arguments: map[string]string{
				"component": "object storage",
				"id_prefix": "",
			},
			wantErr:     true,
			errContains: "id_prefix",
		},
		{
			name: "id prefix and component embedded in YAML template",
			arguments: map[string]string{
				"component": "message queue",
				"id_prefix": "SEC.SLAM.MQ",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				text := result.Messages[2].Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "SEC.SLAM.MQ")
				assert.Contains(t, text, "message queue")
				assert.Contains(t, text, "id: SEC.SLAM.MQ")
				assert.Contains(t, text, "message queue Security Threat Catalog")
				assert.Contains(t, text, "SEC.SLAM.MQ.CAP##")
				assert.Contains(t, text, "SEC.SLAM.MQ.THR##")
			},
		},
		{
			name: "component with control characters rejected",
			arguments: map[string]string{
				"component": "bad\x00value",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with control characters rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME\nPLAT",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "component with markdown link injection rejected",
			arguments: map[string]string{
				"component": "click](http://evil.com)",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "component with html comment rejected",
			arguments: map[string]string{
				"component": "<!-- override -->",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "component with template interpolation rejected",
			arguments: map[string]string{
				"component": "${EVIL_VAR}",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "component with backticks rejected",
			arguments: map[string]string{
				"component": "name`; drop table",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with lowercase letters rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "acme.plat.cr",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with spaces rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME PLAT CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with template interpolation rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.${INJECT}",
			},
			wantErr:     true,
			errContains: "must match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.GetPromptRequest{
				Params: &mcp.GetPromptParams{
					Name:      "threat_assessment",
					Arguments: tt.arguments,
				},
			}

			result, err := handler(ctx, req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestNewThreatAssessmentHandlerLexiconFetchError(t *testing.T) {
	handler := NewThreatAssessmentHandler(failingLexiconFetcher, mockSchemaFetcher)
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "threat_assessment",
			Arguments: map[string]string{"component": "test", "id_prefix": "ACME.TEST"},
		},
	}

	_, err := handler(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching lexicon")
}

func TestNewThreatAssessmentHandlerSchemaFetchError(t *testing.T) {
	handler := NewThreatAssessmentHandler(mockLexiconFetcher, failingSchemaFetcher)
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "threat_assessment",
			Arguments: map[string]string{"component": "test", "id_prefix": "ACME.TEST"},
		},
	}

	_, err := handler(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching schema docs")
	assert.Contains(t, err.Error(), "network error")
}

func TestNewControlCatalogHandler(t *testing.T) {
	handler := NewControlCatalogHandler(mockLexiconFetcher, mockSchemaFetcher)

	tests := []struct {
		name           string
		arguments      map[string]string
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, result *mcp.GetPromptResult)
	}{
		{
			name: "successful prompt generation with embedded resources",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "container runtime")
				assert.Contains(t, result.Description, "ACME.PLAT.CR")
				require.Len(t, result.Messages, 5, "2 embedded resources + 3 text messages")

				assertEmbeddedResources(t, result.Messages)

				instructionMsg := result.Messages[2]
				assert.Equal(t, mcp.Role("user"), instructionMsg.Role)
				text := instructionMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "container runtime")
				assert.Contains(t, text, "ACME.PLAT.CR")
				assert.Contains(t, text, "## Embedded Resources")
				assert.Contains(t, text, "## Available Tool")
				assert.Contains(t, text, "**Catalog Import**")
				assert.Contains(t, text, "**Scope and Metadata**")
				assert.Contains(t, text, "**Define Control Families**")
				assert.Contains(t, text, "**Define Controls**")
				assert.Contains(t, text, "**Assemble and Validate**")
				assert.Contains(t, text, "**Next Steps**")
				assert.Contains(t, text, "validate_gemara_artifact")
				assert.Contains(t, text, "Privateer")
				assert.Contains(t, text, "#ControlCatalog")

				assistantMsg := result.Messages[3]
				assert.Equal(t, mcp.Role("assistant"), assistantMsg.Role)
				assistantText := assistantMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, assistantText, "container runtime")
				assert.Contains(t, assistantText, "FINOS CCC Core")
				assert.Contains(t, assistantText, "Step 1: Catalog Import")
				assert.Contains(t, assistantText, "reply \"yes\"")

				userMsg := result.Messages[4]
				assert.Equal(t, mcp.Role("user"), userMsg.Role)
				userText := userMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, userText, "container runtime")
				assert.Contains(t, userText, "ACME.PLAT.CR")
			},
		},
		{
			name: "component with dots hyphens underscores accepted",
			arguments: map[string]string{
				"component": "kube-apiserver_v2.1",
				"id_prefix": "ACME.PLAT.KA",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "kube-apiserver_v2.1")
			},
		},
		{
			name: "missing component argument",
			arguments: map[string]string{
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "missing id_prefix argument",
			arguments: map[string]string{
				"component": "API gateway",
			},
			wantErr:     true,
			errContains: "id_prefix",
		},
		{
			name:        "both arguments missing",
			arguments:   map[string]string{},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "empty component string",
			arguments: map[string]string{
				"component": "",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "empty id_prefix string",
			arguments: map[string]string{
				"component": "object storage",
				"id_prefix": "",
			},
			wantErr:     true,
			errContains: "id_prefix",
		},
		{
			name: "id prefix and component embedded in YAML template",
			arguments: map[string]string{
				"component": "message queue",
				"id_prefix": "SEC.SLAM.MQ",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				text := result.Messages[2].Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "SEC.SLAM.MQ")
				assert.Contains(t, text, "message queue")
				assert.Contains(t, text, "id: SEC.SLAM.MQ")
				assert.Contains(t, text, "message queue Security Control Catalog")
				assert.Contains(t, text, "SEC.SLAM.MQ.C##")
				assert.Contains(t, text, "SEC.SLAM.MQ.C##.TR##")
			},
		},
		{
			name: "component with control characters rejected",
			arguments: map[string]string{
				"component": "bad\x00value",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "component with markdown link injection rejected",
			arguments: map[string]string{
				"component": "click](http://evil.com)",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "component with html tags rejected",
			arguments: map[string]string{
				"component": "<script>alert(1)</script>",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "component with template interpolation rejected",
			arguments: map[string]string{
				"component": "${EVIL_VAR}",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with template interpolation rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.${INJECT}",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with lowercase letters rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "acme.plat.cr",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with spaces rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME PLAT CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with underscores rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME_PLAT_CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.GetPromptRequest{
				Params: &mcp.GetPromptParams{
					Name:      "control_catalog",
					Arguments: tt.arguments,
				},
			}

			result, err := handler(ctx, req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestNewControlCatalogHandlerLexiconFetchError(t *testing.T) {
	handler := NewControlCatalogHandler(failingLexiconFetcher, mockSchemaFetcher)
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "control_catalog",
			Arguments: map[string]string{"component": "test", "id_prefix": "ACME.TEST"},
		},
	}

	_, err := handler(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching lexicon")
}

func TestNewControlCatalogHandlerSchemaFetchError(t *testing.T) {
	handler := NewControlCatalogHandler(mockLexiconFetcher, failingSchemaFetcher)
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "control_catalog",
			Arguments: map[string]string{"component": "test", "id_prefix": "ACME.TEST"},
		},
	}

	_, err := handler(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching schema docs")
	assert.Contains(t, err.Error(), "network error")
}

func TestPromptControlCatalogMetadata(t *testing.T) {
	assert.Equal(t, "control_catalog", PromptControlCatalog.Name)
	assert.NotEmpty(t, PromptControlCatalog.Description)
	require.Len(t, PromptControlCatalog.Arguments, 2)

	componentArg := PromptControlCatalog.Arguments[0]
	assert.Equal(t, "component", componentArg.Name)
	assert.True(t, componentArg.Required)

	prefixArg := PromptControlCatalog.Arguments[1]
	assert.Equal(t, "id_prefix", prefixArg.Name)
	assert.True(t, prefixArg.Required)
}

func TestPromptThreatAssessmentMetadata(t *testing.T) {
	assert.Equal(t, "threat_assessment", PromptThreatAssessment.Name)
	assert.NotEmpty(t, PromptThreatAssessment.Description)
	require.Len(t, PromptThreatAssessment.Arguments, 2)

	componentArg := PromptThreatAssessment.Arguments[0]
	assert.Equal(t, "component", componentArg.Name)
	assert.True(t, componentArg.Required)

	prefixArg := PromptThreatAssessment.Arguments[1]
	assert.Equal(t, "id_prefix", prefixArg.Name)
	assert.True(t, prefixArg.Required)
}

func TestNewPolicyHandler(t *testing.T) {
	handler := NewPolicyHandler(mockLexiconFetcher, mockSchemaFetcher)

	tests := []struct {
		name           string
		arguments      map[string]string
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, result *mcp.GetPromptResult)
	}{
		{
			name: "successful prompt generation with embedded resources",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "container runtime")
				assert.Contains(t, result.Description, "ACME.PLAT.CR")
				require.Len(t, result.Messages, 5, "2 embedded resources + 3 text messages")

				assertEmbeddedResources(t, result.Messages)

				instructionMsg := result.Messages[2]
				assert.Equal(t, mcp.Role("user"), instructionMsg.Role)
				text := instructionMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "container runtime")
				assert.Contains(t, text, "ACME.PLAT.CR")
				assert.Contains(t, text, "## Embedded Resources")
				assert.Contains(t, text, "## Available Tool")
				assert.Contains(t, text, "**Catalog and Artifact Import**")
				assert.Contains(t, text, "**Scope and Metadata**")
				assert.Contains(t, text, "**Define Contacts (RACI)**")
				assert.Contains(t, text, "**Define Scope**")
				assert.Contains(t, text, "**Define Imports**")
				assert.Contains(t, text, "**Define Adherence**")
				assert.Contains(t, text, "**Assemble and Validate**")
				assert.Contains(t, text, "**Next Steps**")
				assert.Contains(t, text, "validate_gemara_artifact")
				assert.Contains(t, text, "#Policy")
				assert.Contains(t, text, `gemara-version: "1.0.0-rc.0"`)
				assert.Contains(t, text, "github.com/gemaraproj/gemara@v1.0.0-rc.0")

				assistantMsg := result.Messages[3]
				assert.Equal(t, mcp.Role("assistant"), assistantMsg.Role)
				assistantText := assistantMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, assistantText, "container runtime")
				assert.Contains(t, assistantText, "Step 1: Catalog and Artifact Import")
				assert.Contains(t, assistantText, "Control Catalog")

				userMsg := result.Messages[4]
				assert.Equal(t, mcp.Role("user"), userMsg.Role)
				userText := userMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, userText, "container runtime")
				assert.Contains(t, userText, "ACME.PLAT.CR")
			},
		},
		{
			name: "component with dots hyphens underscores accepted",
			arguments: map[string]string{
				"component": "kube-apiserver_v2.1",
				"id_prefix": "ACME.PLAT.KA",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "kube-apiserver_v2.1")
			},
		},
		{
			name: "missing component argument",
			arguments: map[string]string{
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "missing id_prefix argument",
			arguments: map[string]string{
				"component": "API gateway",
			},
			wantErr:     true,
			errContains: "id_prefix",
		},
		{
			name:        "both arguments missing",
			arguments:   map[string]string{},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "empty component string",
			arguments: map[string]string{
				"component": "",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "empty id_prefix string",
			arguments: map[string]string{
				"component": "object storage",
				"id_prefix": "",
			},
			wantErr:     true,
			errContains: "id_prefix",
		},
		{
			name: "id prefix and component embedded in YAML template",
			arguments: map[string]string{
				"component": "message queue",
				"id_prefix": "SEC.SLAM.MQ",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				text := result.Messages[2].Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "SEC.SLAM.MQ")
				assert.Contains(t, text, "message queue")
				assert.Contains(t, text, "id: SEC.SLAM.MQ")
				assert.Contains(t, text, "message queue Security Policy")
				assert.Contains(t, text, "SEC.SLAM.MQ.CON##")
				assert.Contains(t, text, "SEC.SLAM.MQ.AP##")
			},
		},
		{
			name: "component with control characters rejected",
			arguments: map[string]string{
				"component": "bad\x00value",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "component with template interpolation rejected",
			arguments: map[string]string{
				"component": "${EVIL_VAR}",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with lowercase letters rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "acme.plat.cr",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with spaces rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME PLAT CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with template interpolation rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.${INJECT}",
			},
			wantErr:     true,
			errContains: "must match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.GetPromptRequest{
				Params: &mcp.GetPromptParams{
					Name:      "policy",
					Arguments: tt.arguments,
				},
			}

			result, err := handler(ctx, req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestNewPolicyHandlerLexiconFetchError(t *testing.T) {
	handler := NewPolicyHandler(failingLexiconFetcher, mockSchemaFetcher)
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "policy",
			Arguments: map[string]string{"component": "test", "id_prefix": "ACME.TEST"},
		},
	}

	_, err := handler(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching lexicon")
}

func TestNewPolicyHandlerSchemaFetchError(t *testing.T) {
	handler := NewPolicyHandler(mockLexiconFetcher, failingSchemaFetcher)
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "policy",
			Arguments: map[string]string{"component": "test", "id_prefix": "ACME.TEST"},
		},
	}

	_, err := handler(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching schema docs")
	assert.Contains(t, err.Error(), "network error")
}

func TestPromptPolicyMetadata(t *testing.T) {
	assert.Equal(t, "policy", PromptPolicy.Name)
	assert.NotEmpty(t, PromptPolicy.Description)
	require.Len(t, PromptPolicy.Arguments, 2)

	componentArg := PromptPolicy.Arguments[0]
	assert.Equal(t, "component", componentArg.Name)
	assert.True(t, componentArg.Required)

	prefixArg := PromptPolicy.Arguments[1]
	assert.Equal(t, "id_prefix", prefixArg.Name)
	assert.True(t, prefixArg.Required)
}

func TestNewRiskCatalogHandler(t *testing.T) {
	handler := NewRiskCatalogHandler(mockLexiconFetcher, mockSchemaFetcher)

	tests := []struct {
		name           string
		arguments      map[string]string
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, result *mcp.GetPromptResult)
	}{
		{
			name: "successful prompt generation with embedded resources",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "container runtime")
				assert.Contains(t, result.Description, "ACME.PLAT.CR")
				require.Len(t, result.Messages, 5, "2 embedded resources + 3 text messages")

				assertEmbeddedResources(t, result.Messages)

				instructionMsg := result.Messages[2]
				assert.Equal(t, mcp.Role("user"), instructionMsg.Role)
				text := instructionMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "container runtime")
				assert.Contains(t, text, "ACME.PLAT.CR")
				assert.Contains(t, text, "## Embedded Resources")
				assert.Contains(t, text, "## Available Tool")
				assert.Contains(t, text, "**Threat Catalog Import**")
				assert.Contains(t, text, "**Scope and Metadata**")
				assert.Contains(t, text, "**Define Risk Categories**")
				assert.Contains(t, text, "**Define Risks**")
				assert.Contains(t, text, "**Assemble and Validate**")
				assert.Contains(t, text, "**Next Steps**")
				assert.Contains(t, text, "validate_gemara_artifact")
				assert.Contains(t, text, "#RiskCatalog")
				assert.Contains(t, text, `gemara-version: "1.0.0-rc.0"`)
				assert.Contains(t, text, "github.com/gemaraproj/gemara@v1.0.0-rc.0")

				assistantMsg := result.Messages[3]
				assert.Equal(t, mcp.Role("assistant"), assistantMsg.Role)
				assistantText := assistantMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, assistantText, "container runtime")
				assert.Contains(t, assistantText, "Step 1: Threat Catalog Import")
				assert.Contains(t, assistantText, "Threat Catalog")

				userMsg := result.Messages[4]
				assert.Equal(t, mcp.Role("user"), userMsg.Role)
				userText := userMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, userText, "container runtime")
				assert.Contains(t, userText, "ACME.PLAT.CR")
			},
		},
		{
			name: "component with dots hyphens underscores accepted",
			arguments: map[string]string{
				"component": "kube-apiserver_v2.1",
				"id_prefix": "ACME.PLAT.KA",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "kube-apiserver_v2.1")
			},
		},
		{
			name: "missing component argument",
			arguments: map[string]string{
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "missing id_prefix argument",
			arguments: map[string]string{
				"component": "API gateway",
			},
			wantErr:     true,
			errContains: "id_prefix",
		},
		{
			name:        "both arguments missing",
			arguments:   map[string]string{},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "empty component string",
			arguments: map[string]string{
				"component": "",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "component",
		},
		{
			name: "empty id_prefix string",
			arguments: map[string]string{
				"component": "object storage",
				"id_prefix": "",
			},
			wantErr:     true,
			errContains: "id_prefix",
		},
		{
			name: "id prefix and component embedded in YAML template",
			arguments: map[string]string{
				"component": "message queue",
				"id_prefix": "SEC.SLAM.MQ",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				text := result.Messages[2].Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "SEC.SLAM.MQ")
				assert.Contains(t, text, "message queue")
				assert.Contains(t, text, "id: SEC.SLAM.MQ")
				assert.Contains(t, text, "message queue Risk Catalog")
				assert.Contains(t, text, "SEC.SLAM.MQ.RSK##")
			},
		},
		{
			name: "component with control characters rejected",
			arguments: map[string]string{
				"component": "bad\x00value",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "component with template interpolation rejected",
			arguments: map[string]string{
				"component": "${EVIL_VAR}",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with lowercase letters rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "acme.plat.cr",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with spaces rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME PLAT CR",
			},
			wantErr:     true,
			errContains: "must match",
		},
		{
			name: "id_prefix with template interpolation rejected",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.${INJECT}",
			},
			wantErr:     true,
			errContains: "must match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.GetPromptRequest{
				Params: &mcp.GetPromptParams{
					Name:      "risk_catalog",
					Arguments: tt.arguments,
				},
			}

			result, err := handler(ctx, req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestNewRiskCatalogHandlerLexiconFetchError(t *testing.T) {
	handler := NewRiskCatalogHandler(failingLexiconFetcher, mockSchemaFetcher)
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "risk_catalog",
			Arguments: map[string]string{"component": "test", "id_prefix": "ACME.TEST"},
		},
	}

	_, err := handler(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching lexicon")
}

func TestNewRiskCatalogHandlerSchemaFetchError(t *testing.T) {
	handler := NewRiskCatalogHandler(mockLexiconFetcher, failingSchemaFetcher)
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "risk_catalog",
			Arguments: map[string]string{"component": "test", "id_prefix": "ACME.TEST"},
		},
	}

	_, err := handler(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching schema docs")
	assert.Contains(t, err.Error(), "network error")
}

func TestPromptRiskCatalogMetadata(t *testing.T) {
	assert.Equal(t, "risk_catalog", PromptRiskCatalog.Name)
	assert.NotEmpty(t, PromptRiskCatalog.Description)
	require.Len(t, PromptRiskCatalog.Arguments, 2)

	componentArg := PromptRiskCatalog.Arguments[0]
	assert.Equal(t, "component", componentArg.Name)
	assert.True(t, componentArg.Required)

	prefixArg := PromptRiskCatalog.Arguments[1]
	assert.Equal(t, "id_prefix", prefixArg.Name)
	assert.True(t, prefixArg.Required)
}
