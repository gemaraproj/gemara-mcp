// SPDX-License-Identifier: Apache-2.0

package evidence_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gemaraproj/gemara-mcp/internal/evidence"
	"github.com/gemaraproj/gemara-mcp/internal/evidence/parsers"
)

// ---------------------------------------------------------------------------
// SchemaMapper
// ---------------------------------------------------------------------------

func TestSchemaMapper_Map(t *testing.T) {
	mapper := evidence.NewSchemaMapper()

	tests := []struct {
		name            string
		chunks          []evidence.EvidenceChunk
		wantMinCount    int
		wantTargetField string // at least one candidate should map to this field
	}{
		{
			name: "objective keyword maps to controls objective",
			chunks: []evidence.EvidenceChunk{
				{Text: "The objective of this control is to ensure TLS 1.2+", SourceID: "doc.md", SectionPath: "Section 1", Confidence: 1.0},
			},
			wantMinCount:    1,
			wantTargetField: "controls[].objective",
		},
		{
			name: "title keyword maps to metadata title",
			chunks: []evidence.EvidenceChunk{
				{Text: "title: Network Security Policy", SourceID: "policy.yaml", SectionPath: "root", Confidence: 1.0},
			},
			wantMinCount:    1,
			wantTargetField: "metadata.title",
		},
		{
			name: "assessment keyword maps to controls assessment",
			chunks: []evidence.EvidenceChunk{
				{Text: "Verify that TLS certificates are valid and unexpired", SourceID: "doc.md", SectionPath: "Audit", Confidence: 0.9},
			},
			wantMinCount:    1,
			wantTargetField: "controls[].assessment",
		},
		{
			name:         "unrecognised text produces no candidates",
			chunks:       []evidence.EvidenceChunk{{Text: "Lorem ipsum dolor sit amet", SourceID: "doc.md", SectionPath: "random", Confidence: 1.0}},
			wantMinCount: 0,
		},
		{
			name:         "empty chunks list returns empty candidates",
			chunks:       []evidence.EvidenceChunk{},
			wantMinCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := mapper.Map(tt.chunks)
			assert.GreaterOrEqual(t, len(candidates), tt.wantMinCount)

			if tt.wantTargetField != "" {
				found := false
				for _, c := range candidates {
					if c.TargetField == tt.wantTargetField {
						found = true
						break
					}
				}
				assert.True(t, found, "expected at least one candidate with TargetField=%q, got: %+v", tt.wantTargetField, candidates)
			}
		})
	}
}

func TestSchemaMapper_ConfidencePropagation(t *testing.T) {
	mapper := evidence.NewSchemaMapper()
	chunks := []evidence.EvidenceChunk{
		{Text: "objective: ensure encryption", SourceID: "doc.md", SectionPath: "s1", Confidence: 1.0},
		{Text: "objective: ensure encryption", SourceID: "doc.md", SectionPath: "s2", Confidence: 0.5},
	}
	candidates := mapper.Map(chunks)
	require.Len(t, candidates, 2)
	assert.Greater(t, candidates[0].Confidence, candidates[1].Confidence, "higher chunk confidence should yield higher candidate confidence")
}

// ---------------------------------------------------------------------------
// Pipeline
// ---------------------------------------------------------------------------

func TestPipeline_UnsupportedFormat(t *testing.T) {
	p := evidence.NewPipeline() // no parsers registered
	_, err := p.Run(context.Background(), evidence.EvidenceSource{
		Content: []byte("anything"),
		Format:  "pdf",
		ID:      "test.pdf",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported evidence format")
}

func TestPipeline_RegisteredParsers(t *testing.T) {
	p := evidence.NewPipeline(parsers.NewMarkdownParser(), parsers.NewYAMLParser())
	names := p.RegisteredParsers()
	assert.Equal(t, []string{"markdown", "yaml"}, names)
}

func TestPipeline_RunWithMeta_MarkdownDoc(t *testing.T) {
	p := evidence.NewPipeline(parsers.NewMarkdownParser())
	src := evidence.EvidenceSource{
		Content: []byte("# Network Security\nThe objective of this control is to encrypt all traffic.\n\n## Assessment\nVerify TLS settings."),
		ID:      "policy.md",
	}
	result, err := p.RunWithMeta(context.Background(), src)
	require.NoError(t, err)
	assert.Equal(t, "markdown", result.ParserUsed)
	assert.Greater(t, result.ChunkCount, 0)
}

// ---------------------------------------------------------------------------
// MarkdownParser
// ---------------------------------------------------------------------------

func TestMarkdownParser_CanHandle(t *testing.T) {
	p := parsers.NewMarkdownParser()

	assert.True(t, p.CanHandle(evidence.EvidenceSource{Format: "markdown"}))
	assert.True(t, p.CanHandle(evidence.EvidenceSource{Format: "md"}))
	assert.True(t, p.CanHandle(evidence.EvidenceSource{Content: []byte("# Heading\ntext")}))
	assert.True(t, p.CanHandle(evidence.EvidenceSource{Content: []byte("preamble\n# Heading")}))
	assert.False(t, p.CanHandle(evidence.EvidenceSource{Content: []byte("apiVersion: v1")}))
}

func TestMarkdownParser_Parse(t *testing.T) {
	p := parsers.NewMarkdownParser()
	src := evidence.EvidenceSource{
		Content: []byte("# Section One\nContent of section one.\n\n## Subsection\nMore content here.\n\n# Section Two\nAnother section."),
		ID:      "test.md",
	}
	chunks, err := p.Parse(context.Background(), src)
	require.NoError(t, err)
	assert.Len(t, chunks, 3) // Section One, Subsection, Section Two

	assert.Equal(t, "Section One", chunks[0].SectionPath)
	assert.Contains(t, chunks[0].Text, "Content of section one")
	assert.Equal(t, "test.md", chunks[0].SourceID)
	assert.Equal(t, 0.85, chunks[0].Confidence)
}

func TestMarkdownParser_Parse_Preamble(t *testing.T) {
	p := parsers.NewMarkdownParser()
	src := evidence.EvidenceSource{
		Content: []byte("This is a preamble.\n\n# First Section\nSection content."),
		ID:      "doc.md",
	}
	chunks, err := p.Parse(context.Background(), src)
	require.NoError(t, err)
	assert.Len(t, chunks, 2)
	assert.Equal(t, "preamble", chunks[0].SectionPath)
}

func TestMarkdownParser_Parse_EmptyContent(t *testing.T) {
	p := parsers.NewMarkdownParser()
	chunks, err := p.Parse(context.Background(), evidence.EvidenceSource{Content: []byte(""), ID: "empty.md"})
	require.NoError(t, err)
	assert.Empty(t, chunks)
}

// ---------------------------------------------------------------------------
// YAMLParser
// ---------------------------------------------------------------------------

func TestYAMLParser_CanHandle(t *testing.T) {
	p := parsers.NewYAMLParser()

	assert.True(t, p.CanHandle(evidence.EvidenceSource{Format: "yaml"}))
	assert.True(t, p.CanHandle(evidence.EvidenceSource{Format: "yml"}))
	assert.True(t, p.CanHandle(evidence.EvidenceSource{Format: "json"}))
	assert.True(t, p.CanHandle(evidence.EvidenceSource{Content: []byte(`{"key": "value"}`)}))
	assert.True(t, p.CanHandle(evidence.EvidenceSource{Content: []byte("key: value\nother: thing")}))
	// Should NOT steal Markdown content
	assert.False(t, p.CanHandle(evidence.EvidenceSource{Content: []byte("# Heading\ntext")}))
}

func TestYAMLParser_Parse(t *testing.T) {
	p := parsers.NewYAMLParser()
	src := evidence.EvidenceSource{
		Content: []byte("title: My Policy\nversion: \"1.0\"\nobjective: Ensure security"),
		ID:      "policy.yaml",
	}
	chunks, err := p.Parse(context.Background(), src)
	require.NoError(t, err)
	assert.NotEmpty(t, chunks)

	// All chunks should have the correct SourceID
	for _, c := range chunks {
		assert.Equal(t, "policy.yaml", c.SourceID)
		assert.Equal(t, 0.80, c.Confidence)
	}
}

func TestYAMLParser_Parse_InvalidYAML(t *testing.T) {
	p := parsers.NewYAMLParser()
	_, err := p.Parse(context.Background(), evidence.EvidenceSource{
		Content: []byte("invalid: [unclosed"),
		ID:      "bad.yaml",
	})
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// KubernetesParser
// ---------------------------------------------------------------------------

const sampleDeployment = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  securityContext:
    runAsNonRoot: true
  containers:
    - name: app
      image: my-app:1.0
      env:
        - name: SECRET
          value: "abc"
`

func TestKubernetesParser_CanHandle(t *testing.T) {
	p := parsers.NewKubernetesParser()

	assert.True(t, p.CanHandle(evidence.EvidenceSource{Format: "kubernetes"}))
	assert.True(t, p.CanHandle(evidence.EvidenceSource{Format: "k8s"}))
	assert.True(t, p.CanHandle(evidence.EvidenceSource{Content: []byte(sampleDeployment)}))
	assert.False(t, p.CanHandle(evidence.EvidenceSource{Content: []byte("# Markdown doc")}))
}

func TestKubernetesParser_Parse(t *testing.T) {
	p := parsers.NewKubernetesParser()
	src := evidence.EvidenceSource{Content: []byte(sampleDeployment), ID: "deploy.yaml"}
	chunks, err := p.Parse(context.Background(), src)
	require.NoError(t, err)
	assert.NotEmpty(t, chunks)

	// Should have extracted at least the identity chunk and some spec chunks
	var sectionPaths []string
	for _, c := range chunks {
		sectionPaths = append(sectionPaths, c.SectionPath)
	}

	hasIdentity := false
	hasSecurityCtx := false
	for _, sp := range sectionPaths {
		if contains(sp, "identity") {
			hasIdentity = true
		}
		if contains(sp, "securityContext") {
			hasSecurityCtx = true
		}
	}
	assert.True(t, hasIdentity, "should have an identity chunk")
	assert.True(t, hasSecurityCtx, "should have a securityContext chunk")
}

func TestKubernetesParser_Parse_MultiDoc(t *testing.T) {
	p := parsers.NewKubernetesParser()
	content := sampleDeployment + "\n---\napiVersion: v1\nkind: Service\nmetadata:\n  name: my-svc\nspec:\n  type: ClusterIP\n"
	src := evidence.EvidenceSource{Content: []byte(content), ID: "multi.yaml"}
	chunks, err := p.Parse(context.Background(), src)
	require.NoError(t, err)

	kinds := map[string]bool{}
	for _, c := range chunks {
		if contains(c.Text, "kind: Deployment") {
			kinds["Deployment"] = true
		}
		if contains(c.Text, "kind: Service") {
			kinds["Service"] = true
		}
	}
	assert.True(t, kinds["Deployment"], "should parse Deployment document")
	assert.True(t, kinds["Service"], "should parse Service document")
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
