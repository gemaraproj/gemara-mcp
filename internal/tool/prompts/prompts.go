// SPDX-License-Identifier: Apache-2.0

package prompts

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/gemaraproj/gemara-mcp/internal/tool/consts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	validComponentPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 ._-]*$`)
	validIDPrefixPattern  = regexp.MustCompile(`^[A-Z0-9.-]+$`)
)

const maxPromptArgLen = 200

// LexiconFetcher retrieves the lexicon content at prompt invocation time.
type LexiconFetcher func(ctx context.Context) (string, error)

// SchemaDocsFetcher retrieves formatted schema documentation at prompt invocation time.
// This allows version-specific schema content to be resolved per-session.
type SchemaDocsFetcher func(ctx context.Context) (string, error)

func validateComponent(value string) error {
	if value == "" {
		return fmt.Errorf("component argument is required")
	}
	if len(value) > maxPromptArgLen {
		return fmt.Errorf("component argument exceeds maximum length of %d", maxPromptArgLen)
	}
	if !validComponentPattern.MatchString(value) {
		return fmt.Errorf("component %q must match ^[a-zA-Z0-9][a-zA-Z0-9 ._-]*$ (letters, digits, spaces, dots, underscores, hyphens)", value)
	}
	return nil
}

func validateIDPrefix(value string) error {
	if value == "" {
		return fmt.Errorf("id_prefix argument is required")
	}
	if len(value) > maxPromptArgLen {
		return fmt.Errorf("id_prefix argument exceeds maximum length of %d", maxPromptArgLen)
	}
	if !validIDPrefixPattern.MatchString(value) {
		return fmt.Errorf("id_prefix %q must match ^[A-Z0-9.-]+$ (uppercase letters, digits, dots, hyphens only)", value)
	}
	return nil
}

func embeddedResourceMessages(lexicon string, schemaDocs string) []*mcp.PromptMessage {
	return []*mcp.PromptMessage{
		{
			Role: "user",
			Content: &mcp.EmbeddedResource{
				Resource: &mcp.ResourceContents{
					URI:      consts.LexiconResourceURI,
					MIMEType: "text/yaml",
					Text:     lexicon,
				},
			},
		},
		{
			Role: "user",
			Content: &mcp.EmbeddedResource{
				Resource: &mcp.ResourceContents{
					URI:      consts.SchemaDocsResourceURI,
					MIMEType: "text/plain",
					Text:     schemaDocs,
				},
			},
		},
	}
}

var (
	//go:embed threat_assessment_system.md
	threatAssessmentSystemTemplate string

	//go:embed threat_assessment_assistant.md
	threatAssessmentAssistantTemplate string

	//go:embed threat_assessment_user.md
	threatAssessmentUserTemplate string

	//go:embed control_catalog_system.md
	controlCatalogSystemTemplate string

	//go:embed control_catalog_assistant.md
	controlCatalogAssistantTemplate string

	//go:embed control_catalog_user.md
	controlCatalogUserTemplate string
)

// PromptThreatAssessment is the MCP prompt definition for the threat assessment wizard.
var PromptThreatAssessment = &mcp.Prompt{
	Name:        "threat_assessment",
	Title:       "Threat Assessment Wizard",
	Description: "Interactive wizard that guides you through creating a Gemara-compatible Threat Catalog (Layer 2) for your project.",
	Arguments: []*mcp.PromptArgument{
		{
			Name:        "component",
			Title:       "Component Name",
			Description: "The name of the component or technology to assess (e.g., 'container runtime', 'API gateway', 'object storage')",
			Required:    true,
		},
		{
			Name:        "id_prefix",
			Title:       "ID Prefix",
			Description: "Organization and project prefix for identifiers in ORG.PROJECT.COMPONENT format (e.g., 'ACME.PLAT.GW')",
			Required:    true,
		},
	},
}

// PromptControlCatalog is the MCP prompt definition for the control catalog wizard.
var PromptControlCatalog = &mcp.Prompt{
	Name:        "control_catalog",
	Title:       "Control Catalog Wizard",
	Description: "Interactive wizard that guides you through creating a Gemara-compatible Control Catalog (Layer 2) for your project.",
	Arguments: []*mcp.PromptArgument{
		{
			Name:        "component",
			Title:       "Component Name",
			Description: "The name of the component or technology to create controls for (e.g., 'container runtime', 'API gateway', 'object storage')",
			Required:    true,
		},
		{
			Name:        "id_prefix",
			Title:       "ID Prefix",
			Description: "Organization and project prefix for identifiers in ORG.PROJECT.COMPONENT format (e.g., 'ACME.PLAT.GW')",
			Required:    true,
		},
	},
}

// NewControlCatalogHandler returns a PromptHandler that embeds the lexicon and schema
// docs as EmbeddedResource messages, guaranteeing the LLM receives both during the wizard.
func NewControlCatalogHandler(fetchLexicon LexiconFetcher, fetchSchemaDocs SchemaDocsFetcher) mcp.PromptHandler {
	return func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		if req.Params == nil || req.Params.Arguments == nil {
			return nil, fmt.Errorf("component argument is required")
		}

		component := req.Params.Arguments["component"]
		idPrefix := req.Params.Arguments["id_prefix"]

		if err := validateComponent(component); err != nil {
			return nil, err
		}
		if err := validateIDPrefix(idPrefix); err != nil {
			return nil, err
		}

		lexicon, err := fetchLexicon(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetching lexicon: %w", err)
		}

		schemaDocs, err := fetchSchemaDocs(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetching schema docs: %w", err)
		}

		r := strings.NewReplacer("${COMPONENT}", component, "${ID_PREFIX}", idPrefix)
		resources := embeddedResourceMessages(lexicon, schemaDocs)

		messages := make([]*mcp.PromptMessage, 0, len(resources)+3)
		messages = append(messages, resources...)
		messages = append(messages,
			&mcp.PromptMessage{
				Role:    "user",
				Content: &mcp.TextContent{Text: r.Replace(controlCatalogSystemTemplate)},
			},
			&mcp.PromptMessage{
				Role:    "assistant",
				Content: &mcp.TextContent{Text: r.Replace(controlCatalogAssistantTemplate)},
			},
			&mcp.PromptMessage{
				Role:    "user",
				Content: &mcp.TextContent{Text: r.Replace(controlCatalogUserTemplate)},
			},
		)

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Control catalog wizard for %s (%s)", component, idPrefix),
			Messages:    messages,
		}, nil
	}
}

// NewThreatAssessmentHandler returns a PromptHandler that embeds the lexicon and schema
// docs as EmbeddedResource messages, guaranteeing the LLM receives both during the wizard.
func NewThreatAssessmentHandler(fetchLexicon LexiconFetcher, fetchSchemaDocs SchemaDocsFetcher) mcp.PromptHandler {
	return func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		if req.Params == nil || req.Params.Arguments == nil {
			return nil, fmt.Errorf("component argument is required")
		}

		component := req.Params.Arguments["component"]
		idPrefix := req.Params.Arguments["id_prefix"]

		if err := validateComponent(component); err != nil {
			return nil, err
		}
		if err := validateIDPrefix(idPrefix); err != nil {
			return nil, err
		}

		lexicon, err := fetchLexicon(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetching lexicon: %w", err)
		}

		schemaDocs, err := fetchSchemaDocs(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetching schema docs: %w", err)
		}

		r := strings.NewReplacer("${COMPONENT}", component, "${ID_PREFIX}", idPrefix)
		resources := embeddedResourceMessages(lexicon, schemaDocs)

		messages := make([]*mcp.PromptMessage, 0, len(resources)+3)
		messages = append(messages, resources...)
		messages = append(messages,
			&mcp.PromptMessage{
				Role:    "user",
				Content: &mcp.TextContent{Text: r.Replace(threatAssessmentSystemTemplate)},
			},
			&mcp.PromptMessage{
				Role:    "assistant",
				Content: &mcp.TextContent{Text: r.Replace(threatAssessmentAssistantTemplate)},
			},
			&mcp.PromptMessage{
				Role:    "user",
				Content: &mcp.TextContent{Text: r.Replace(threatAssessmentUserTemplate)},
			},
		)

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Threat assessment wizard for %s (%s)", component, idPrefix),
			Messages:    messages,
		}, nil
	}
}
