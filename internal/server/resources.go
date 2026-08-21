// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
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

//go:embed about.md
var EmbeddedAbout string

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

var ResourceAbout = &mcp.Resource{
	URI:         AboutResourceURI,
	Name:        "gemara-about",
	Title:       "About Gemara",
	Description: "LLM-optimized overview of Gemara: the seven-layer GRC model, artifact types, schemas, and SDKs. Read this before fetching Gemara documentation from the web.",
	MIMEType:    "text/markdown",
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

	// Validate before caching (see validatingFetcher) so an invalid 200
	// response never poisons the cache for the TTL.
	vf := validatingFetcher{inner: hf, validate: func(data []byte) error {
		_, err := parseLexicon(data)
		return err
	}}
	cf := fetcher.NewCachedFetcher[[]byte](vf, a.lexiconCache, hf.URL())
	data, src, err := cf.Fetch(ctx, false)
	if err != nil {
		slog.Warn("failed to fetch lexicon, using embedded fallback", "error", err)
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

func (a *AdvisoryMode) handleAboutResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	content, source := a.fetchAbout(ctx)
	slog.Info("about resource read", "source", source, "size", len(content))
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     content,
		}},
	}, nil
}

// fetchAbout retrieves the LLM-optimized guidance from the upstream URL,
// falling back to the embedded copy when the remote is unreachable, returns a
// non-200, or fails a basic sanity check.
func (a *AdvisoryMode) fetchAbout(ctx context.Context) (content string, source string) {
	hf, err := fetcher.NewStaticHTTPFetcher(aboutURL)
	if err != nil {
		slog.Warn("failed to build about fetch URL, using embedded fallback", "error", err)
		return EmbeddedAbout, "embedded"
	}

	// Validate before caching so an invalid 200 response (e.g. a proxy error
	// page) is never stored: caching it would keep failing validation on every
	// cache hit and pin the resource to the embedded copy for the whole TTL,
	// even after the upstream recovers.
	vf := validatingFetcher{inner: hf, validate: func(data []byte) error {
		if !isValidAbout(data) {
			return fmt.Errorf("content is not a valid Gemara about document")
		}
		return nil
	}}
	cf := fetcher.NewCachedFetcher[[]byte](vf, a.aboutCache, hf.URL())
	data, src, err := cf.Fetch(ctx, false)
	if err != nil {
		slog.Warn("failed to fetch about, using embedded fallback", "error", err)
		return EmbeddedAbout, "embedded"
	}

	return string(data), src
}

// validatingFetcher wraps a byte fetcher and rejects responses that fail the
// validate func, turning them into fetch errors. Because the error surfaces
// before CachedFetcher stores the result, invalid content never poisons the
// cache and pins the resource to its embedded fallback for the whole TTL.
type validatingFetcher struct {
	inner    fetcher.Fetcher[[]byte]
	validate func([]byte) error
}

func (f validatingFetcher) Fetch(ctx context.Context) ([]byte, string, error) {
	data, src, err := f.inner.Fetch(ctx)
	if err != nil {
		return nil, "", err
	}
	if err := f.validate(data); err != nil {
		return nil, "", fmt.Errorf("fetched content failed validation: %w", err)
	}
	return data, src, nil
}

// utf8BOM is the UTF-8 byte-order mark. bytes.TrimSpace does not strip it, so
// isValidAbout removes it explicitly before inspecting the content.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// isValidAbout performs a minimal sanity check that fetched content is the
// Gemara guidance document and not, for example, an HTML error page. It
// tolerates a leading UTF-8 BOM and variations in the top heading text,
// requiring only that the document starts with a Markdown heading and mentions
// Gemara.
func isValidAbout(data []byte) bool {
	trimmed := bytes.TrimSpace(bytes.TrimPrefix(bytes.TrimSpace(data), utf8BOM))
	if len(trimmed) == 0 {
		return false
	}
	return bytes.HasPrefix(trimmed, []byte("#")) &&
		bytes.Contains(bytes.ToLower(trimmed), []byte("gemara"))
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
