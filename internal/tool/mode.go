// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"

	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
	"github.com/gemaraproj/gemara-mcp/internal/tool/prompts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultSchemaVersion  = "latest"
	defaultLexiconVersion = "v0.19.1"
	lexiconBaseURL        = "https://raw.githubusercontent.com/gemaraproj/gemara/"
	lexiconPathSuffix     = "/docs/lexicon.yaml"
	schemaDocsBaseURL     = "https://registry.cue.works/docs/github.com/gemaraproj/gemara@"
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
	cache                *fetcher.Cache
	lexiconURLBuilder    *fetcher.URLBuilder
	schemaDocsURLBuilder *fetcher.URLBuilder
}

// NewAdvisoryMode creates a new AdvisoryMode with the provided cache and default URLs.
func NewAdvisoryMode(cache *fetcher.Cache) (*AdvisoryMode, error) {
	lexBuilder, err := fetcher.NewURLBuilder(lexiconBaseURL, lexiconPathSuffix)
	if err != nil {
		return nil, fmt.Errorf("configuring lexicon URL: %w", err)
	}
	schemaBuilder, err := fetcher.NewURLBuilder(schemaDocsBaseURL, "")
	if err != nil {
		return nil, fmt.Errorf("configuring schema docs URL: %w", err)
	}
	return &AdvisoryMode{
		cache:                cache,
		lexiconURLBuilder:    lexBuilder,
		schemaDocsURLBuilder: schemaBuilder,
	}, nil
}

func (a *AdvisoryMode) Name() string {
	return "advisory"
}

func (a *AdvisoryMode) Description() string {
	return "Advisory mode: Provides information about Gemara artifacts in the workspace (read-only)"
}

func (a *AdvisoryMode) Register(server *mcp.Server) {
	mcp.AddTool(server, MetadataGetLexicon, a.getLexicon)
	mcp.AddTool(server, MetadataValidateGemaraArtifact, ValidateGemaraArtifact)
	mcp.AddTool(server, MetadataGetSchemaDocs, a.getSchemaDocs)
}

// ArtifactMode extends AdvisoryMode with guided wizards for creating Gemara artifacts.
type ArtifactMode struct {
	*AdvisoryMode
}

// NewArtifactMode creates a new ArtifactMode with all AdvisoryMode capabilities plus artifact prompts.
func NewArtifactMode(cache *fetcher.Cache) (*ArtifactMode, error) {
	advisory, err := NewAdvisoryMode(cache)
	if err != nil {
		return nil, err
	}
	return &ArtifactMode{AdvisoryMode: advisory}, nil
}

func (a *ArtifactMode) Name() string {
	return "artifact"
}

func (a *ArtifactMode) Description() string {
	return "Artifact mode: Guides for creating Gemara-compatible security artifacts (includes all advisory capabilities)"
}

func (a *ArtifactMode) Register(server *mcp.Server) {
	a.AdvisoryMode.Register(server)
	server.AddPrompt(prompts.PromptThreatAssessment, prompts.HandleThreatAssessment)
	server.AddPrompt(prompts.PromptControlCatalog, prompts.HandleControlCatalog)
}

// getLexicon wraps GetLexicon with cache access and configuration.
func (a *AdvisoryMode) getLexicon(ctx context.Context, req *mcp.CallToolRequest, input InputGetLexicon) (*mcp.CallToolResult, OutputGetLexicon, error) {
	version := input.Version
	if version == "" {
		version = defaultLexiconVersion
	}
	f, err := fetcher.NewHTTPFetcher(a.lexiconURLBuilder, version)
	if err != nil {
		return nil, OutputGetLexicon{}, err
	}
	cf := fetcher.NewCachedFetcher(f, a.cache, f.URL())
	return GetLexicon(ctx, req, input, cf)
}

// getSchemaDocs wraps GetSchemaDocs with cache access and configuration.
func (a *AdvisoryMode) getSchemaDocs(ctx context.Context, req *mcp.CallToolRequest, input InputGetSchemaDocs) (*mcp.CallToolResult, OutputGetSchemaDocs, error) {
	version := input.Version
	if version == "" {
		version = defaultSchemaVersion
	}
	f, err := fetcher.NewHTTPFetcher(a.schemaDocsURLBuilder, version)
	if err != nil {
		return nil, OutputGetSchemaDocs{}, err
	}
	cf := fetcher.NewCachedFetcher(f, a.cache, f.URL())
	return GetSchemaDocs(ctx, req, input, cf)
}
