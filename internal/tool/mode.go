// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
)

const (
	httpTimeout          = 30 * time.Second
	defaultSchemaVersion = "latest"
)

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
type AdvisoryMode struct {
	cache             *fetcher.Cache
	lexiconURL        string
	schemaDocsBaseURL string
}

// NewAdvisoryMode creates a new AdvisoryMode with the provided cache and default URLs.
func NewAdvisoryMode(cache *fetcher.Cache) *AdvisoryMode {
	return &AdvisoryMode{
		cache:             cache,
		lexiconURL:        "https://raw.githubusercontent.com/gemaraproj/gemara/main/docs/lexicon.yaml",
		schemaDocsBaseURL: "https://registry.cue.works/docs/github.com/gemaraproj/gemara@",
	}
}

func (a AdvisoryMode) Name() string {
	return "advisory"
}

func (a AdvisoryMode) Description() string {
	return "Advisory mode: Provides information about Gemara artifacts in the workspace (read-only)"
}

func (a AdvisoryMode) Register(server *mcp.Server) {
	// Lexicon tool - provides information about Gemara terms
	mcp.AddTool(server, MetadataGetLexicon, a.getLexicon)

	// Validation tool - validates artifacts without modifying them
	mcp.AddTool(server, MetadataValidateGemaraArtifact, ValidateGemaraArtifact)

	// Schema documentation tool - retrieves schema documentation from CUE registry
	mcp.AddTool(server, MetadataGetSchemaDocs, a.getSchemaDocs)
}

// getLexicon wraps GetLexicon with cache access and configuration.
func (a AdvisoryMode) getLexicon(ctx context.Context, req *mcp.CallToolRequest, input InputGetLexicon) (*mcp.CallToolResult, OutputGetLexicon, error) {
	source := a.lexiconURL
	f := fetcher.NewHTTPFetcher(source, httpTimeout)
	cf := fetcher.NewCachedFetcher(f, a.cache, source)
	return GetLexicon(ctx, req, input, cf)
}

// getSchemaDocs wraps GetSchemaDocs with cache access and configuration.
func (a AdvisoryMode) getSchemaDocs(ctx context.Context, req *mcp.CallToolRequest, input InputGetSchemaDocs) (*mcp.CallToolResult, OutputGetSchemaDocs, error) {
	version := input.Version
	if version == "" {
		version = defaultSchemaVersion
	}
	source := a.schemaDocsBaseURL + version
	f := fetcher.NewHTTPFetcher(source, httpTimeout)
	cf := fetcher.NewCachedFetcher(f, a.cache, source)
	return GetSchemaDocs(ctx, req, input, cf)
}
