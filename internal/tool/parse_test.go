// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGovernanceDocument(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name           string
		input          InputParseGovernanceDocument
		wantErr        bool
		errContains    string
		validateOutput func(t *testing.T, output OutputParseGovernanceDocument)
	}{
		{
			name:        "empty content returns error",
			input:       InputParseGovernanceDocument{Content: ""},
			wantErr:     true,
			errContains: "content is required",
		},
		{
			name: "markdown governance document produces candidates",
			input: InputParseGovernanceDocument{
				Content:  "# Network Security\nThe objective of this control is to encrypt all traffic.\n\n## Assessment\nVerify that TLS 1.2 or higher is enforced on all endpoints.",
				Format:   "markdown",
				SourceID: "network-policy.md",
			},
			wantErr: false,
			validateOutput: func(t *testing.T, output OutputParseGovernanceDocument) {
				assert.Equal(t, "markdown", output.ParserUsed)
				assert.Greater(t, output.TotalChunks, 0, "should extract at least one chunk")
				assert.NotEmpty(t, output.Candidates, "should produce at least one candidate")
				for _, c := range output.Candidates {
					assert.NotEmpty(t, c.TargetField, "candidate must have a target field")
					assert.NotEmpty(t, c.Value, "candidate must have a value")
					assert.Greater(t, c.Confidence, 0.0, "candidate confidence must be positive")
					assert.LessOrEqual(t, c.Confidence, 1.0, "candidate confidence must be <= 1.0")
				}
			},
		},
		{
			name: "kubernetes manifest produces candidates",
			input: InputParseGovernanceDocument{
				Content: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: secure-app
spec:
  securityContext:
    runAsNonRoot: true
  containers:
    - name: app
      image: myapp:1.0
`,
				Format:   "kubernetes",
				SourceID: "deployment.yaml",
			},
			wantErr: false,
			validateOutput: func(t *testing.T, output OutputParseGovernanceDocument) {
				assert.Equal(t, "kubernetes", output.ParserUsed)
				assert.Greater(t, output.TotalChunks, 0)
			},
		},
		{
			name: "yaml config produces candidates",
			input: InputParseGovernanceDocument{
				Content:  "title: Data Encryption Policy\nversion: \"2.0\"\nobjective: Ensure all data at rest is encrypted",
				Format:   "yaml",
				SourceID: "config.yaml",
			},
			wantErr: false,
			validateOutput: func(t *testing.T, output OutputParseGovernanceDocument) {
				assert.Equal(t, "yaml", output.ParserUsed)
				assert.NotEmpty(t, output.Candidates)
			},
		},
		{
			name: "auto-detection without format hint",
			input: InputParseGovernanceDocument{
				Content: "# Auto-detected Markdown\nThis is a control. The objective is to ensure compliance.",
				// No Format field — auto-detection should kick in
			},
			wantErr: false,
			validateOutput: func(t *testing.T, output OutputParseGovernanceDocument) {
				assert.NotEmpty(t, output.ParserUsed)
				assert.Greater(t, output.TotalChunks, 0)
			},
		},
		{
			name: "source_id is optional and defaults gracefully",
			input: InputParseGovernanceDocument{
				Content: "# Policy\nThe objective of this policy is to enforce access controls.",
				Format:  "markdown",
				// No SourceID
			},
			wantErr: false,
			validateOutput: func(t *testing.T, output OutputParseGovernanceDocument) {
				assert.NotEmpty(t, output.Candidates)
				// SourceRef should still be populated
				for _, c := range output.Candidates {
					assert.NotEmpty(t, c.SourceRef)
				}
			},
		},
		{
			name: "unsupported format returns error",
			input: InputParseGovernanceDocument{
				Content: "some binary or unsupported content",
				Format:  "pdf",
			},
			wantErr:     true,
			errContains: "unsupported evidence format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := ParseGovernanceDocument(ctx, req, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}
