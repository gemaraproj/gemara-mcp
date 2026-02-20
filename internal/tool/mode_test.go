// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"
	"time"

	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var advisoryToolNames = []string{
	"get_lexicon",
	"validate_gemara_artifact",
	"get_schema_docs",
}

var artifactPromptNames = []string{
	"threat_assessment",
	"control_catalog",
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

func TestAdvisoryModeRegistersToolsOnly(t *testing.T) {
	cache := fetcher.NewCache(1 * time.Hour)
	mode, err := NewAdvisoryMode(cache)
	require.NoError(t, err)
	server := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.0"},
		&mcp.ServerOptions{Instructions: mode.Description()},
	)
	mode.Register(server)

	session := connectSession(t, server)
	tools := toolNames(t, session)
	prompts := promptNames(t, session)

	for _, name := range advisoryToolNames {
		assert.Contains(t, tools, name)
	}
	for _, name := range artifactPromptNames {
		assert.NotContains(t, prompts, name, "advisory mode must not register artifact prompts")
	}
}

func TestArtifactModeRegistersToolsAndPrompts(t *testing.T) {
	cache := fetcher.NewCache(1 * time.Hour)
	mode, err := NewArtifactMode(cache)
	require.NoError(t, err)
	server := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.0"},
		&mcp.ServerOptions{Instructions: mode.Description()},
	)
	mode.Register(server)

	session := connectSession(t, server)
	tools := toolNames(t, session)
	prompts := promptNames(t, session)

	for _, name := range advisoryToolNames {
		assert.Contains(t, tools, name, "artifact mode must include all advisory tools")
	}
	for _, name := range artifactPromptNames {
		assert.Contains(t, prompts, name, "artifact mode must register artifact prompts")
	}
}

func TestAdvisoryModeMetadata(t *testing.T) {
	cache := fetcher.NewCache(1 * time.Hour)
	mode, err := NewAdvisoryMode(cache)
	require.NoError(t, err)
	assert.Equal(t, "advisory", mode.Name())
	assert.Contains(t, mode.Description(), "read-only")
	assert.Contains(t, mode.Description(), "not create new ones")
}

func TestArtifactModeMetadata(t *testing.T) {
	cache := fetcher.NewCache(1 * time.Hour)
	mode, err := NewArtifactMode(cache)
	require.NoError(t, err)
	assert.Equal(t, "artifact", mode.Name())
	assert.Contains(t, mode.Description(), "advisory capabilities")
	assert.Contains(t, mode.Description(), "Orient every response toward creation")
}

func TestModeInterfaceCompliance(t *testing.T) {
	cache := fetcher.NewCache(1 * time.Hour)

	advisory, err := NewAdvisoryMode(cache)
	require.NoError(t, err)
	var _ Mode = advisory

	artifact, err := NewArtifactMode(cache)
	require.NoError(t, err)
	var _ Mode = artifact
}
