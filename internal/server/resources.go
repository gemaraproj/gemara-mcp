// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/url"

	"cuelang.org/go/cue"
	gemara "github.com/gemaraproj/go-gemara"
	goyaml "github.com/goccy/go-yaml"

	"github.com/gemaraproj/gemara-mcp/internal/server/fetcher"
	"github.com/gemaraproj/gemara-mcp/internal/server/schema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed lexicon.yaml
var EmbeddedLexicon string

var ResourceLexicon = &mcp.Resource{
	URI:         LexiconResourceURI,
	Name:        "gemara-lexicon",
	Title:       "Gemara Lexicon",
	Description: "Term definitions for the Gemara security model.",
	MIMEType:    "text/yaml",
	Annotations: &mcp.Annotations{
		Audience: []mcp.Role{mcp.Role("assistant")},
		Priority: 1,
	},
}

var ResourceSchemaDocs = &mcp.Resource{
	URI:         SchemaDocsResourceURI,
	Name:        "gemara-schema-docs",
	Title:       "Gemara Schema Documentation",
	Description: "CUE schema definitions for all Gemara artifact types (latest version). Use the versioned resource template for a specific version.",
	MIMEType:    "text/plain",
	Annotations: &mcp.Annotations{
		Audience: []mcp.Role{mcp.Role("assistant")},
		Priority: 1,
	},
}

var ResourceSchemaDocsTemplate = &mcp.ResourceTemplate{
	URITemplate: SchemaDocsResourceURITemplate,
	Name:        "gemara-schema-docs-versioned",
	Title:       "Gemara Schema Documentation (versioned)",
	Description: "CUE schema definitions for a specific Gemara module version. Accepts a semver version parameter (e.g., v1.2.3) or 'latest'.",
	MIMEType:    "text/plain",
}

func (a *AdvisoryMode) handleLexiconResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	content, source := a.fetchLexicon(ctx)
	slog.Info("lexicon resource read", "source", source, "size", len(content))
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/yaml",
			Text:     content,
		}},
	}, nil
}

// parseLexicon validates YAML content against the gemara.Lexicon type from
// the SDK. Uses goccy/go-yaml directly because the SDK's codec package is
// internal and gemara.Load assumes a file/URL fetch workflow.
func parseLexicon(data []byte) (*gemara.Lexicon, error) {
	var lex gemara.Lexicon
	if err := goyaml.Unmarshal(data, &lex); err != nil {
		return nil, fmt.Errorf("unmarshalling lexicon: %w", err)
	}
	if lex.Title == "" {
		return nil, fmt.Errorf("lexicon missing required field: title")
	}
	if len(lex.Terms) == 0 {
		return nil, fmt.Errorf("lexicon contains no terms")
	}
	return &lex, nil
}

// fetchLexicon retrieves the lexicon from the remote URL, falling back to
// the embedded copy on failure.
func (a *AdvisoryMode) fetchLexicon(ctx context.Context) (content string, source string) {
	version, err := a.resolveLexiconVersion(ctx)
	if err != nil {
		slog.Warn("failed to resolve lexicon version, using embedded fallback", "error", err)
		return EmbeddedLexicon, "embedded"
	}

	hf, err := fetcher.NewHTTPFetcher(a.lexiconURLBuilder, version)
	if err != nil {
		slog.Warn("failed to build lexicon fetch URL, using embedded fallback", "error", err)
		return EmbeddedLexicon, "embedded"
	}

	cf := fetcher.NewCachedFetcher[[]byte](hf, a.lexiconCache, hf.URL())
	data, src, err := cf.Fetch(ctx, false)
	if err != nil {
		slog.Warn("failed to fetch lexicon, using embedded fallback", "error", err)
		return EmbeddedLexicon, "embedded"
	}

	if _, err := parseLexicon(data); err != nil {
		slog.Warn("remote lexicon failed validation, using embedded fallback", "error", err)
		return EmbeddedLexicon, "embedded"
	}

	return string(data), src
}

// resolveLexiconVersion resolves "latest" to a concrete semver tag via
// the CUE module registry.
func (a *AdvisoryMode) resolveLexiconVersion(ctx context.Context) (string, error) {
	tag, _, err := a.versionResolver.Fetch(ctx, false)
	if err != nil {
		return "", fmt.Errorf("resolving latest version: %w", err)
	}
	return tag, nil
}

func (a *AdvisoryMode) handleSchemaDocsResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return a.fetchSchemaDocsForVersion(ctx, req.Params.URI, defaultSchemaVersion)
}

func (a *AdvisoryMode) handleSchemaDocsTemplateResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	version, err := parseSchemaDocsVersion(req.Params.URI)
	if err != nil {
		return nil, err
	}
	return a.fetchSchemaDocsForVersion(ctx, req.Params.URI, version)
}

func (a *AdvisoryMode) fetchSchemaDocsForVersion(ctx context.Context, uri, version string) (*mcp.ReadResourceResult, error) {
	modulePath := gemaraModuleBase + version
	f := schema.NewCUERegistryFetcher(modulePath)
	cf := fetcher.NewCachedFetcher[cue.Value](f, a.schemaCache, modulePath)

	val, source, err := cf.Fetch(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema: %w", err)
	}

	defs, err := schema.FormatDefinitions(val)
	if err != nil {
		return nil, fmt.Errorf("failed to format schema: %w", err)
	}

	slog.Info("schema docs resource read", "version", version, "source", source, "size", len(defs))
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "text/plain",
			Text:     defs,
		}},
	}, nil
}

// parseSchemaDocsVersion extracts and validates the version query parameter
// from a schema docs resource URI, defaulting to "latest" when absent.
func parseSchemaDocsVersion(rawURI string) (string, error) {
	u, err := url.Parse(rawURI)
	if err != nil {
		return "", fmt.Errorf("invalid resource URI: %w", err)
	}
	version := u.Query().Get("version")
	if version == "" {
		return defaultSchemaVersion, nil
	}
	if err := fetcher.ValidateVersion(version); err != nil {
		return "", err
	}
	return version, nil
}
