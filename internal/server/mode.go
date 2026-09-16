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

// gemaraPreamble is a static overview prepended to every mode's instructions.
// It gives the client enough context to answer basic "what is Gemara" questions
// without fetching from the open web, and points to gemara://about for detail.
const gemaraPreamble = `Gemara is a seven-layer conceptual model for Governance, Risk, and Compliance (GRC) engineering, plus CUE schemas and SDKs that make GRC artifacts interoperable. It is part of the OpenSSF ecosystem.

The seven layers (each builds on the one below):
1. Vectors & Guidance   — generic, high-level risk guidance (e.g., OWASP Top 10, NIST CSF)
2. Threats & Controls   — technology-specific, threat-informed controls with assessment requirements
3. Risk & Policy        — organization-specific risk catalogs and policies
4. Sensitive Activities — the pivot point: the actions that require governance
5. Evaluation           — opinions formed by inspecting policy compliance (intent and behavior)
6. Enforcement          — preventive and remediative action on non-compliance findings
7. Audit & Continuous Monitoring — point-in-time review and ongoing observation

Layers 1-3 (Definition) produce Catalogs; layers 5-7 (Measurement) produce Logs.
Artifact types include: Guidance/Vector/Principle Catalogs (L1); Control/Capability/Threat Catalogs (L2); Risk Catalog and Policy (L3); Evaluation/Enforcement/Audit Logs (L5-7); plus Mapping Documents and the Lexicon.

For a deeper overview, read the gemara://about resource before fetching Gemara docs from the web. Use gemara://lexicon for term definitions and gemara://schema/definitions for the CUE schemas.

`

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
	aboutCache        *fetcher.Cache[[]byte]
	versionResolver   *fetcher.CachedFetcher[string]
	lexiconURLBuilder *fetcher.URLBuilder
}

// NewAdvisoryMode creates a new AdvisoryMode with the provided cache TTL.
func NewAdvisoryMode(cacheTTL time.Duration) (*AdvisoryMode, error) {
	if _, err := parseLexicon([]byte(EmbeddedLexicon)); err != nil {
		return nil, fmt.Errorf("embedded lexicon is invalid: %w", err)
	}
	if !isValidAbout([]byte(EmbeddedAbout)) {
		return nil, fmt.Errorf("embedded about content is invalid")
	}

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
		aboutCache:        fetcher.NewCache[[]byte](cacheTTL),
		versionResolver:   versionResolver,
		lexiconURLBuilder: lexiconBuilder,
	}, nil
}

func (a *AdvisoryMode) Name() string {
	return "advisory"
}

func (a *AdvisoryMode) Description() string {
	return gemaraPreamble + `Gemara advisory mode. Analyze and validate existing security artifacts.

Tools: validate_gemara_artifact. Resources: gemara://about, gemara://lexicon, gemara://schema/definitions. Resource templates: gemara://schema/definitions{?version}.

For artifact creation, suggest switching to artifact mode.`
}

func (a *AdvisoryMode) Register(server *mcp.Server) {
	mcp.AddTool(server, MetadataValidateGemaraArtifact, a.validateGemaraArtifact)
	server.AddResource(ResourceAbout, a.handleAboutResource)
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
	return gemaraPreamble + `Gemara artifact mode. Create, iterate on, and validate security artifacts.

Tools: validate_gemara_artifact, migrate_gemara_artifact. Resources: gemara://about, gemara://lexicon, gemara://schema/definitions. Resource templates: gemara://schema/definitions{?version}. Prompts: threat_assessment, control_catalog, migration.

Offer wizard prompts for new artifacts. Validate frequently during iteration.`
}

func (a *ArtifactMode) Register(server *mcp.Server) {
	a.AdvisoryMode.Register(server)

	mcp.AddTool(server, MetadataMigrateGemaraArtifact, a.migrateGemaraArtifact)

	fetchLexicon := a.lexiconFetcher()
	fetchSchemaDocs := a.schemaDocsFetcher()
	server.AddPrompt(PromptThreatAssessment, NewThreatAssessmentHandler(fetchLexicon, fetchSchemaDocs))
	server.AddPrompt(PromptControlCatalog, NewControlCatalogHandler(fetchLexicon, fetchSchemaDocs))
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
	if err := fetcher.ValidateVersion(version); err != nil {
		return nil, OutputValidateGemaraArtifact{}, err
	}
	modulePath := gemaraModuleBase + version
	f := schema.NewCUERegistryFetcher(modulePath)
	cf := fetcher.NewCachedFetcher[cue.Value](f, a.schemaCache, modulePath)
	return ValidateGemaraArtifact(ctx, req, input, cf)
}
