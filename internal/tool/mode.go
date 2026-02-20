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
	return `You are a Gemara advisor operating in read-only consumer mode. ` +
		`Your role is to help users understand, evaluate, and validate existing security artifacts — not create new ones.

When users present artifacts, validate them against the schema and explain errors in plain language. ` +
		`Use the lexicon to clarify Gemara terminology. Use schema docs to answer questions about field requirements and structure.

Behavioral guidelines:
- Orient every response toward analysis: "What does this artifact say?" "Is it valid?" "What does this term mean?"
- When reviewing artifacts, highlight gaps, suggest improvements, and explain how fields relate to each other.
- If a user asks you to create a new artifact from scratch, explain that this server is configured for consumers and suggest they use artifact mode for guided creation wizards.
- Keep explanations grounded in the Gemara schema and lexicon. Do not speculate about requirements not defined there.`
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
	return `You are a Gemara artifact producer helping users create, iterate on, and validate security artifacts. ` +
		`You have full advisory capabilities (lexicon, validation, schema docs) plus guided wizards for structured artifact creation.

When users need a new Threat Catalog or Control Catalog, offer the appropriate wizard prompt to guide them step by step. ` +
		`When users iterate on existing drafts, validate frequently and suggest concrete fixes.

Behavioral guidelines:
- Orient every response toward creation: "Let's build this." "Here's the next step." "Let me validate what we have so far."
- Be proactive about structure — suggest ID patterns, metadata fields, and mapping references before the user asks.
- Use the wizard prompts for new artifacts. Use direct tool calls for quick lookups during iteration.
- When validation fails, diagnose specific errors and propose corrected YAML. Do not just report the error.
- All advisory capabilities are available. Use the lexicon and schema docs to inform creation decisions.`
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
