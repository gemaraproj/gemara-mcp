// SPDX-License-Identifier: Apache-2.0

package tool

import "github.com/modelcontextprotocol/go-sdk/mcp"

// Mode represents the operational mode of the MCP server.
type Mode interface {
	// Name returns the string representation of the mode.
	Name() string
	// Description returns a human-readable description of the mode.
	Description() string
	// Register adds mode-related tools to the mcp server
	Register(*mcp.Server)
}

// AdvisoryMode defines tools and resources for operating in a read-only query mode
type AdvisoryMode struct{}

func (a AdvisoryMode) Name() string {
	return "advisory"
}

func (a AdvisoryMode) Description() string {
	return "Advisory mode: Provides information about Gemara artifacts in the workspace (read-only)"
}

func (a AdvisoryMode) Register(server *mcp.Server) {
	// Lexicon tool - provides information about Gemara terms
	server.AddResource(MetadataLexiconResource, HandleLexiconResource)
	server.AddResource(MetadataLexiconResourceAlias, HandleLexiconResource)
	mcp.AddTool(server, MetadataGetLexicon, GetLexicon)

	// Validation tool - validates artifacts without modifying them
	mcp.AddTool(server, MetadataValidateGemaraArtifact, ValidateGemaraArtifact)
}
