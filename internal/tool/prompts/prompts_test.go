// SPDX-License-Identifier: Apache-2.0

package prompts

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleThreatAssessment(t *testing.T) {
	tests := []struct {
		name           string
		arguments      map[string]string
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, result *mcp.GetPromptResult)
	}{
		{
			name: "successful prompt generation",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "container runtime")
				assert.Contains(t, result.Description, "ACME.PLAT.CR")
				require.Len(t, result.Messages, 2, "should have two messages")

				assistantMsg := result.Messages[0]
				assert.Equal(t, mcp.Role("assistant"), assistantMsg.Role)
				text := assistantMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "container runtime")
				assert.Contains(t, text, "ACME.PLAT.CR")
				assert.Contains(t, text, "Phase 0")
				assert.Contains(t, text, "Phase 1")
				assert.Contains(t, text, "Phase 2")
				assert.Contains(t, text, "Phase 3")
				assert.Contains(t, text, "Phase 4")
				assert.Contains(t, text, "Phase 5")
				assert.Contains(t, text, "validate_gemara_artifact")
				assert.Contains(t, text, "FINOS Common Cloud Controls")
				assert.Contains(t, text, "Privateer")

				userMsg := result.Messages[1]
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
				text := result.Messages[0].Content.(*mcp.TextContent).Text
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

			result, err := HandleThreatAssessment(ctx, req)

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

func TestHandleControlCatalog(t *testing.T) {
	tests := []struct {
		name           string
		arguments      map[string]string
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, result *mcp.GetPromptResult)
	}{
		{
			name: "successful prompt generation",
			arguments: map[string]string{
				"component": "container runtime",
				"id_prefix": "ACME.PLAT.CR",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *mcp.GetPromptResult) {
				assert.Contains(t, result.Description, "container runtime")
				assert.Contains(t, result.Description, "ACME.PLAT.CR")
				require.Len(t, result.Messages, 2, "should have two messages")

				assistantMsg := result.Messages[0]
				assert.Equal(t, mcp.Role("assistant"), assistantMsg.Role)
				text := assistantMsg.Content.(*mcp.TextContent).Text
				assert.Contains(t, text, "container runtime")
				assert.Contains(t, text, "ACME.PLAT.CR")
				assert.Contains(t, text, "Phase 0")
				assert.Contains(t, text, "Phase 1")
				assert.Contains(t, text, "Phase 2")
				assert.Contains(t, text, "Phase 3")
				assert.Contains(t, text, "Phase 4")
				assert.Contains(t, text, "Phase 5")
				assert.Contains(t, text, "validate_gemara_artifact")
				assert.Contains(t, text, "FINOS Common Cloud Controls")
				assert.Contains(t, text, "Privateer")
				assert.Contains(t, text, "#ControlCatalog")

				userMsg := result.Messages[1]
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
				text := result.Messages[0].Content.(*mcp.TextContent).Text
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

			result, err := HandleControlCatalog(ctx, req)

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
