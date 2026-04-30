// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cuelang.org/go/cue"
	"github.com/gemaraproj/gemara-mcp/internal/server/fetcher"
	"github.com/gemaraproj/gemara-mcp/internal/server/schema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultSchemaVersion = "latest"
	gemaraModulePath     = "github.com/gemaraproj/gemara"
	gemaraModuleBase     = gemaraModulePath + "@"
	lexiconBaseURL       = "https://raw.githubusercontent.com/gemaraproj/gemara/"
	lexiconPathSuffix    = "/docs/lexicon.yaml"
)

// Mode represents the operational mode of the MCP server.
type Mode interface {
	// Name returns the string representation of the mode.
	Name() string
	// Description returns a human-readable description of the mode.
	Description() string
	// Register adds mode-related tools and resources to the mcp server
	Register(*mcp.Server)
}

// AdvisoryMode defines tools and resources for operating in a read-only query mode
type AdvisoryMode struct {
	schemaCache       *fetcher.Cache[cue.Value]
	lexiconCache      *fetcher.Cache[[]byte]
	versionResolver   *fetcher.CachedFetcher[string]
	lexiconURLBuilder *fetcher.URLBuilder
}

// NewAdvisoryMode creates a new AdvisoryMode with the provided cache TTL.
func NewAdvisoryMode(cacheTTL time.Duration) (*AdvisoryMode, error) {
	lexiconBuilder, err := fetcher.NewURLBuilder(lexiconBaseURL, lexiconPathSuffix)
	if err != nil {
		return nil, fmt.Errorf("creating lexicon URL builder: %w", err)
	}
	versionCache := fetcher.NewCache[string](cacheTTL)
	resolver := schema.NewCUEVersionResolver(gemaraModulePath)
	versionResolver := fetcher.NewCachedFetcher[string](resolver, versionCache, gemaraModulePath)

	slog.Info("mode initialized", "mode", "advisory")
	return &AdvisoryMode{
		schemaCache:       fetcher.NewCache[cue.Value](cacheTTL),
		lexiconCache:      fetcher.NewCache[[]byte](cacheTTL),
		versionResolver:   versionResolver,
		lexiconURLBuilder: lexiconBuilder,
	}, nil
}

func (a *AdvisoryMode) Name() string {
	return "advisory"
}

func (a *AdvisoryMode) Description() string {
	return `Gemara advisory mode. Analyze and validate existing security artifacts.

Tools: validate_gemara_artifact. Resources: gemara://lexicon, gemara://schema/definitions. Resource templates: gemara://schema/definitions{?version}.

For artifact creation, suggest switching to artifact mode.`
}

func (a *AdvisoryMode) Register(server *mcp.Server) {
	mcp.AddTool(server, MetadataValidateGemaraArtifact, a.validateGemaraArtifact)
	server.AddResource(ResourceLexicon, a.handleLexiconResource)
	server.AddResource(ResourceSchemaDocs, a.handleSchemaDocsResource)
	server.AddResourceTemplate(ResourceSchemaDocsTemplate, a.handleSchemaDocsTemplateResource)
}

// ArtifactMode extends AdvisoryMode with guided wizards for creating Gemara artifacts.
type ArtifactMode struct {
	*AdvisoryMode
}

// NewArtifactMode creates a new ArtifactMode with all AdvisoryMode capabilities plus artifact prompts.
func NewArtifactMode(cacheTTL time.Duration) (*ArtifactMode, error) {
	advisory, err := NewAdvisoryMode(cacheTTL)
	if err != nil {
		return nil, err
	}
	slog.Info("mode initialized", "mode", "artifact")
	return &ArtifactMode{AdvisoryMode: advisory}, nil
}

func (a *ArtifactMode) Name() string {
	return "artifact"
}

func (a *ArtifactMode) Description() string {
	return `Gemara artifact mode. Create, iterate on, and validate security artifacts.

Tools: validate_gemara_artifact, migrate_gemara_artifact. Resources: gemara://lexicon, gemara://schema/definitions. Resource templates: gemara://schema/definitions{?version}. Prompts: threat_assessment, control_catalog, mapping_document, migration.

Offer wizard prompts for new artifacts. Validate frequently during iteration.`
}

func (a *ArtifactMode) Register(server *mcp.Server) {
	a.AdvisoryMode.Register(server)

	mcp.AddTool(server, MetadataMigrateGemaraArtifact, a.migrateGemaraArtifact)

	fetchLexicon := a.lexiconFetcher()
	fetchSchemaDocs := a.schemaDocsFetcher()
	server.AddPrompt(PromptThreatAssessment, NewThreatAssessmentHandler(fetchLexicon, fetchSchemaDocs))
	server.AddPrompt(PromptControlCatalog, NewControlCatalogHandler(fetchLexicon, fetchSchemaDocs))
	server.AddPrompt(PromptMappingDocument, NewMappingDocumentHandler(fetchLexicon, fetchSchemaDocs))
	server.AddPrompt(PromptMigration, NewMigrationHandler(fetchLexicon, fetchSchemaDocs))
}

func (a *ArtifactMode) migrateGemaraArtifact(ctx context.Context, req *mcp.CallToolRequest, input InputMigrateGemaraArtifact) (*mcp.CallToolResult, OutputMigrateGemaraArtifact, error) {
	return MigrateGemaraArtifact(ctx, req, input)
}

// lexiconFetcher returns a LexiconFetcher that always succeeds because
// fetchLexicon falls back to the embedded lexicon on any remote failure.
func (a *AdvisoryMode) lexiconFetcher() LexiconFetcher {
	return func(ctx context.Context) (content string, source string, err error) {
		content, source = a.fetchLexicon(ctx)
		return content, source, nil
	}
}

func (a *AdvisoryMode) schemaDocsFetcher() SchemaDocsFetcher {
	return func(ctx context.Context) (string, error) {
		modulePath := gemaraModuleBase + defaultSchemaVersion
		f := schema.NewCUERegistryFetcher(modulePath)
		cf := fetcher.NewCachedFetcher[cue.Value](f, a.schemaCache, modulePath)

		val, _, err := cf.Fetch(ctx, false)
		if err != nil {
			return "", fmt.Errorf("failed to fetch schema: %w", err)
		}
		return schema.FormatDefinitions(val)
	}
}

// validateGemaraArtifact wraps ValidateGemaraArtifact with schema cache access.
func (a *AdvisoryMode) validateGemaraArtifact(ctx context.Context, req *mcp.CallToolRequest, input InputValidateGemaraArtifact) (*mcp.CallToolResult, OutputValidateGemaraArtifact, error) {
	version := input.Version
	if version == "" {
		version = defaultSchemaVersion
	}
	modulePath := gemaraModuleBase + version
	f := schema.NewCUERegistryFetcher(modulePath)
	cf := fetcher.NewCachedFetcher[cue.Value](f, a.schemaCache, modulePath)
	return ValidateGemaraArtifact(ctx, req, input, cf)
}
