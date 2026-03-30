// SPDX-License-Identifier: Apache-2.0

//go:build integration

package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadLexiconResource(t *testing.T) {
	session := setupAdvisorySession(t)
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: LexiconResourceURI,
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)

	content := result.Contents[0]
	assert.Equal(t, LexiconResourceURI, content.URI)
	assert.Equal(t, "text/yaml", content.MIMEType)
	assert.Contains(t, content.Text, "term:")
}

func TestReadLexiconResourceReturnsContent(t *testing.T) {
	session := setupAdvisorySession(t)
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: LexiconResourceURI,
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.NotEmpty(t, result.Contents[0].Text, "lexicon should always return content via embedded fallback")
}

func TestReadSchemaDocsResource(t *testing.T) {
	session := setupAdvisorySession(t)
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: SchemaDocsResourceURI,
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)

	content := result.Contents[0]
	assert.Equal(t, SchemaDocsResourceURI, content.URI)
	assert.Equal(t, "text/plain", content.MIMEType)
	assert.NotEmpty(t, content.Text)
}

func TestReadSchemaDocsTemplateResourceLatest(t *testing.T) {
	session := setupAdvisorySession(t)
	uri := "gemara://schema/definitions?version=latest"
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: uri,
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)

	content := result.Contents[0]
	assert.Equal(t, uri, content.URI)
	assert.NotEmpty(t, content.Text)
}

func TestCallValidateGemaraArtifactViaTool(t *testing.T) {
	validContent, err := os.ReadFile(filepath.Join("testdata", "good-ccc.yaml"))
	require.NoError(t, err)

	session := setupAdvisorySession(t)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "validate_gemara_artifact",
		Arguments: map[string]any{
			"artifact_content": string(validContent),
			"definition":       "#ControlCatalog",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestCallValidateGemaraArtifactWithVersion(t *testing.T) {
	validContent, err := os.ReadFile(filepath.Join("testdata", "good-ccc.yaml"))
	require.NoError(t, err)

	session := setupAdvisorySession(t)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "validate_gemara_artifact",
		Arguments: map[string]any{
			"artifact_content": string(validContent),
			"definition":       "#ControlCatalog",
			"version":          "latest",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestArtifactModeResourceAccess(t *testing.T) {
	session := setupArtifactSession(t)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: LexiconResourceURI,
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "term:")
}
