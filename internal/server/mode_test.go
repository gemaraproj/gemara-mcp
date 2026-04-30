// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var advisoryToolNames = []string{
	"validate_gemara_artifact",
}

var artifactToolNames = []string{
	"migrate_gemara_artifact",
}

var advisoryResourceURIs = []string{
	LexiconResourceURI,
	SchemaDocsResourceURI,
}

var advisoryResourceTemplateURIs = []string{
	SchemaDocsResourceURITemplate,
}

var artifactPromptNames = []string{
	"threat_assessment",
	"control_catalog",
	"mapping_document",
	"migration",
}

func connectSession(t *testing.T, server *mcp.Server) *mcp.ClientSession {
	t.Helper()

	ct, st := mcp.NewInMemoryTransports()
	_, err := server.Connect(context.Background(), st, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(context.Background(), ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Logf("failed to close session: %v", err)
		}
	})

	return session
}

func toolNames(t *testing.T, session *mcp.ClientSession) []string {
	t.Helper()
	result, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)
	names := make([]string, len(result.Tools))
	for i, tool := range result.Tools {
		names[i] = tool.Name
	}
	return names
}

func promptNames(t *testing.T, session *mcp.ClientSession) []string {
	t.Helper()
	result, err := session.ListPrompts(context.Background(), nil)
	require.NoError(t, err)
	names := make([]string, len(result.Prompts))
	for i, prompt := range result.Prompts {
		names[i] = prompt.Name
	}
	return names
}

func resourceURIs(t *testing.T, session *mcp.ClientSession) []string {
	t.Helper()
	result, err := session.ListResources(context.Background(), nil)
	require.NoError(t, err)
	uris := make([]string, len(result.Resources))
	for i, r := range result.Resources {
		uris[i] = r.URI
	}
	return uris
}

func resourceTemplateURIs(t *testing.T, session *mcp.ClientSession) []string {
	t.Helper()
	result, err := session.ListResourceTemplates(context.Background(), nil)
	require.NoError(t, err)
	uris := make([]string, len(result.ResourceTemplates))
	for i, rt := range result.ResourceTemplates {
		uris[i] = rt.URITemplate
	}
	return uris
}

func TestAdvisoryModeRegistersToolsAndResources(t *testing.T) {
	mode, err := NewAdvisoryMode(1 * time.Hour)
	require.NoError(t, err)
	server := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.0"},
		&mcp.ServerOptions{Instructions: mode.Description()},
	)
	mode.Register(server)

	session := connectSession(t, server)
	tools := toolNames(t, session)
	resources := resourceURIs(t, session)
	templates := resourceTemplateURIs(t, session)
	prompts := promptNames(t, session)

	for _, name := range advisoryToolNames {
		assert.Contains(t, tools, name)
	}
	assert.NotContains(t, tools, "get_lexicon", "lexicon is now a resource, not a tool")
	assert.NotContains(t, tools, "get_schema_docs", "schema docs is now a resource, not a tool")

	for _, uri := range advisoryResourceURIs {
		assert.Contains(t, resources, uri)
	}

	for _, uri := range advisoryResourceTemplateURIs {
		assert.Contains(t, templates, uri)
	}

	for _, name := range artifactToolNames {
		assert.NotContains(t, tools, name, "advisory mode must not register artifact-only tools")
	}

	for _, name := range artifactPromptNames {
		assert.NotContains(t, prompts, name, "advisory mode must not register artifact prompts")
	}
}

func TestArtifactModeRegistersToolsResourcesAndPrompts(t *testing.T) {
	mode, err := NewArtifactMode(1 * time.Hour)
	require.NoError(t, err)
	server := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.0"},
		&mcp.ServerOptions{Instructions: mode.Description()},
	)
	mode.Register(server)

	session := connectSession(t, server)
	tools := toolNames(t, session)
	resources := resourceURIs(t, session)
	templates := resourceTemplateURIs(t, session)
	prompts := promptNames(t, session)

	for _, name := range advisoryToolNames {
		assert.Contains(t, tools, name, "artifact mode must include all advisory tools")
	}

	for _, name := range artifactToolNames {
		assert.Contains(t, tools, name, "artifact mode must include artifact-only tools")
	}

	for _, uri := range advisoryResourceURIs {
		assert.Contains(t, resources, uri, "artifact mode must include all advisory resources")
	}

	for _, uri := range advisoryResourceTemplateURIs {
		assert.Contains(t, templates, uri, "artifact mode must include all advisory resource templates")
	}

	for _, name := range artifactPromptNames {
		assert.Contains(t, prompts, name, "artifact mode must register artifact prompts")
	}
}

func TestAdvisoryModeMetadata(t *testing.T) {
	mode, err := NewAdvisoryMode(1 * time.Hour)
	require.NoError(t, err)
	assert.Equal(t, "advisory", mode.Name())
	assert.Contains(t, mode.Description(), "advisory mode")
	assert.Contains(t, mode.Description(), "gemara://lexicon")
	assert.Contains(t, mode.Description(), "gemara://schema/definitions{?version}")
	assert.NotContains(t, mode.Description(), "- term:", "lexicon must not be embedded in description")
}

func TestArtifactModeMetadata(t *testing.T) {
	mode, err := NewArtifactMode(1 * time.Hour)
	require.NoError(t, err)
	assert.Equal(t, "artifact", mode.Name())
	assert.Contains(t, mode.Description(), "artifact mode")
	assert.Contains(t, mode.Description(), "validate_gemara_artifact")
	assert.Contains(t, mode.Description(), "migrate_gemara_artifact")
	assert.Contains(t, mode.Description(), "migration")
	assert.Contains(t, mode.Description(), "gemara://lexicon")
	assert.Contains(t, mode.Description(), "gemara://schema/definitions{?version}")
	assert.NotContains(t, mode.Description(), "- term:", "lexicon must not be embedded in description")
}

func TestParseSchemaDocsVersion(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "no version defaults to latest",
			uri:         "gemara://schema/definitions",
			wantVersion: "latest",
		},
		{
			name:        "explicit latest",
			uri:         "gemara://schema/definitions?version=latest",
			wantVersion: "latest",
		},
		{
			name:        "semver version",
			uri:         "gemara://schema/definitions?version=v1.2.3",
			wantVersion: "v1.2.3",
		},
		{
			name:        "semver with prerelease",
			uri:         "gemara://schema/definitions?version=v0.1.0-beta.1",
			wantVersion: "v0.1.0-beta.1",
		},
		{
			name:    "invalid version rejected",
			uri:     "gemara://schema/definitions?version=not-semver",
			wantErr: true,
		},
		{
			name:    "path traversal rejected",
			uri:     "gemara://schema/definitions?version=../../etc/passwd",
			wantErr: true,
		},
		{
			name:        "empty version param defaults to latest",
			uri:         "gemara://schema/definitions?version=",
			wantVersion: "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseSchemaDocsVersion(tt.uri)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantVersion, version)
		})
	}
}

func TestModeInterfaceCompliance(t *testing.T) {
	advisory, err := NewAdvisoryMode(1 * time.Hour)
	require.NoError(t, err)
	var _ Mode = advisory

	artifact, err := NewArtifactMode(1 * time.Hour)
	require.NoError(t, err)
	var _ Mode = artifact
}
