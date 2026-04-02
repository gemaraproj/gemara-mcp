// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func TestReadSchemaDocsTemplateResourceInvalidVersion(t *testing.T) {
	session := setupAdvisorySession(t)
	_, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "gemara://schema/definitions?version=not-semver",
	})
	require.Error(t, err)
}
