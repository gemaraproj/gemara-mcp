// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gemaraproj/gemara-mcp/internal/tool/consts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAdvisorySession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	mode, err := NewAdvisoryMode(1 * time.Hour)
	require.NoError(t, err)
	server := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.0"},
		&mcp.ServerOptions{Instructions: mode.Description()},
	)
	mode.Register(server)
	return connectSession(t, server)
}

func setupArtifactSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	mode, err := NewArtifactMode(1 * time.Hour)
	require.NoError(t, err)
	server := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.0"},
		&mcp.ServerOptions{Instructions: mode.Description()},
	)
	mode.Register(server)
	return connectSession(t, server)
}

func TestReadLexiconResource(t *testing.T) {
	session := setupAdvisorySession(t)
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: consts.LexiconResourceURI,
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)

	content := result.Contents[0]
	assert.Equal(t, consts.LexiconResourceURI, content.URI)
	assert.Equal(t, "text/yaml", content.MIMEType)
	assert.Contains(t, content.Text, "term:")
}

func TestReadLexiconResourceReturnsContent(t *testing.T) {
	session := setupAdvisorySession(t)
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: consts.LexiconResourceURI,
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.NotEmpty(t, result.Contents[0].Text, "lexicon should always return content via embedded fallback")
}

func TestReadSchemaDocsResource(t *testing.T) {
	session := setupAdvisorySession(t)
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: consts.SchemaDocsResourceURI,
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)

	content := result.Contents[0]
	assert.Equal(t, consts.SchemaDocsResourceURI, content.URI)
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

func TestReadSchemaDocsTemplateResourceInvalidVersion(t *testing.T) {
	session := setupAdvisorySession(t)
	_, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "gemara://schema/definitions?version=not-semver",
	})
	require.Error(t, err)
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
		URI: consts.LexiconResourceURI,
	})
	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "term:")
}
