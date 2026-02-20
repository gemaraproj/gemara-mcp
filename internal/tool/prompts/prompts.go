// SPDX-License-Identifier: Apache-2.0

package prompts

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	validComponentPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 ._-]*$`)
	validIDPrefixPattern  = regexp.MustCompile(`^[A-Z0-9.-]+$`)
)

const maxPromptArgLen = 200

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

var (
	//go:embed threat_assessment_system.md
	threatAssessmentSystemTemplate string

	//go:embed threat_assessment_user.md
	threatAssessmentUserTemplate string

	//go:embed control_catalog_system.md
	controlCatalogSystemTemplate string

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

// HandleControlCatalog returns the control catalog wizard prompt messages.
func HandleControlCatalog(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
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

	r := strings.NewReplacer("${COMPONENT}", component, "${ID_PREFIX}", idPrefix)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Control catalog wizard for %s (%s)", component, idPrefix),
		Messages: []*mcp.PromptMessage{
			{
				Role:    "assistant",
				Content: &mcp.TextContent{Text: r.Replace(controlCatalogSystemTemplate)},
			},
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: r.Replace(controlCatalogUserTemplate)},
			},
		},
	}, nil
}

// HandleThreatAssessment returns the threat assessment wizard prompt messages.
func HandleThreatAssessment(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
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

	r := strings.NewReplacer("${COMPONENT}", component, "${ID_PREFIX}", idPrefix)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Threat assessment wizard for %s (%s)", component, idPrefix),
		Messages: []*mcp.PromptMessage{
			{
				Role:    "assistant",
				Content: &mcp.TextContent{Text: r.Replace(threatAssessmentSystemTemplate)},
			},
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: r.Replace(threatAssessmentUserTemplate)},
			},
		},
	}, nil
}
