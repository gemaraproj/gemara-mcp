// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"testing"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateGemaraArtifact(t *testing.T) {
	tests := []struct {
		name           string
		input          InputMigrateGemaraArtifact
		wantErr        bool
		errContains    string
		validateOutput func(t *testing.T, output OutputMigrateGemaraArtifact)
	}{
		{
			name:        "missing artifact_content",
			input:       InputMigrateGemaraArtifact{},
			wantErr:     true,
			errContains: "artifact_content is required",
		},
		{
			name: "invalid YAML",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: "invalid: yaml: [unclosed",
			},
			wantErr:     true,
			errContains: "invalid YAML",
		},
		{
			name: "missing metadata block",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: "title: test\nother: value",
			},
			wantErr:     true,
			errContains: "metadata",
		},
		{
			name: "missing metadata.type",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: "metadata:\n  gemara-version: \"0.20.0\"",
			},
			wantErr:     true,
			errContains: "metadata.type",
		},
		{
			name: "missing metadata.gemara-version",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: "metadata:\n  type: ControlCatalog",
			},
			wantErr:     true,
			errContains: "metadata.gemara-version",
		},
		{
			name: "already at target version rejected",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: "metadata:\n  type: ControlCatalog\n  gemara-version: \"v1.0.0\"",
			},
			wantErr:     true,
			errContains: "already at target",
		},
		{
			name: "unsupported artifact type",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: "metadata:\n  type: Policy\n  gemara-version: \"0.1.0\"",
			},
			wantErr:     true,
			errContains: "unsupported",
		},
		{
			name: "invalid artifact_type rejected",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: "metadata:\n  gemara-version: \"0.20.0\"\ntitle: test\n",
				ArtifactType:    "NotARealType",
			},
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name: "ThreatCatalog migration produces two artifacts",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: testV0ThreatCatalog,
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 2, "ThreatCatalog + CapabilityCatalog")
				assert.NotEmpty(t, output.Changes)
				assert.Contains(t, output.Message, gemara.ThreatCatalogArtifact.String())
				assert.Contains(t, output.Message, "2 artifact(s)")

				tc := output.Artifacts[0]
				assert.Equal(t, gemara.ThreatCatalogArtifact.String(), tc.Type)
				assert.Equal(t, "threats.yaml", tc.SuggestedFilename)
				assert.Contains(t, tc.Content, "gemara-version")
				assert.Contains(t, tc.Content, "v1.0.0")
				assert.NotContains(t, tc.Content, "title: Capability One")
				assert.Contains(t, tc.Content, "mapping-references")
				assert.Contains(t, tc.Content, "TEST.THR01")
				assert.Contains(t, tc.Content, "TEST.CAP01", "referenced by TEST.THR01")
				assert.Contains(t, tc.Content, "Related to cap 1")
				assert.NotContains(t, tc.Content, "title: Capability Two", "inline capabilities removed")

				cc := output.Artifacts[1]
				assert.Equal(t, gemara.CapabilityCatalogArtifact.String(), cc.Type)
				assert.Equal(t, "capabilities.yaml", cc.SuggestedFilename)
				assert.Contains(t, cc.Content, "CapabilityCatalog")
				assert.Contains(t, cc.Content, "v1.0.0")
				assert.Contains(t, cc.Content, "TEST.CAP01")
				assert.Contains(t, cc.Content, "TEST.CAP02")
				assert.Contains(t, cc.Content, "Security Capability Catalog",
					"title should replace 'Threat Catalog' with 'Capability Catalog'")
			},
		},
		{
			name: "ThreatCatalog with imports preserves them",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: testV0ThreatCatalogWithImports,
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 2)
				tc := output.Artifacts[0]
				assert.Contains(t, tc.Content, "imports:", "imports should be preserved in migrated ThreatCatalog")
				assert.Contains(t, tc.Content, "EXT", "import reference-id should be preserved")
				assert.Contains(t, tc.Content, "EXT.THR01", "import entries should be preserved")
				assert.Contains(t, tc.Content, "Imported threat", "import remarks should be preserved")
			},
		},
		{
			name: "ThreatCatalog with nested imports.threats preserves them",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: `metadata:
  id: TEST
  type: ThreatCatalog
  gemara-version: "0.20.0"
  description: Test
  version: 1.0.0
  author:
    id: test
    name: Test
    type: Human
title: Test Security Threat Catalog
imports:
  threats:
    - reference-id: EXT
      entries:
        - reference-id: EXT.THR01
          remarks: Imported threat
threats:
  - id: TEST.THR01
    title: Threat One
    description: First threat
    capabilities:
      - reference-id: TEST
        entries:
          - reference-id: TEST.CAP01
capabilities:
  - id: TEST.CAP01
    title: Cap One
    description: Test
`,
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 2)
				tc := output.Artifacts[0]
				assert.Contains(t, tc.Content, "imports:", "imports should be preserved in migrated ThreatCatalog")
				assert.Contains(t, tc.Content, "EXT", "import reference-id should be preserved")
				assert.Contains(t, tc.Content, "EXT.THR01", "import entries should be preserved")
			},
		},
		{
			name: "ThreatCatalog with imported-threats maps to imports",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: testV0ThreatCatalogWithImportedThreats,
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 2, "ThreatCatalog + CapabilityCatalog")
				tc := output.Artifacts[0]
				assert.Contains(t, tc.Content, "imports:", "imported-threats should be mapped to imports")
				assert.Contains(t, tc.Content, "EXT.THR01", "imported threat entries should be preserved")
				assert.Contains(t, tc.Content, "EXT.THR02", "imported threat entries should be preserved")
				assert.NotContains(t, tc.Content, "imported-threats", "old field name should not appear in output")

				cc := output.Artifacts[1]
				assert.Contains(t, cc.Content, "imports:", "imported-capabilities should be mapped to imports")
				assert.Contains(t, cc.Content, "EXT.CP01", "imported capability entries should be preserved")
				assert.Contains(t, cc.Content, "EXT.CP02", "imported capability entries should be preserved")
				assert.NotContains(t, cc.Content, "imported-capabilities", "old field name should not appear in output")
			},
		},
		{
			name: "ThreatCatalog without capabilities produces one artifact",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: `metadata:
  id: TEST
  type: ThreatCatalog
  gemara-version: "0.20.0"
  description: Test
  version: 1.0.0
  author:
    id: test
    name: Test
    type: Human
title: Test Threat Catalog
threats:
  - id: TEST.THR01
    title: Threat One
    description: First threat
`,
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 1, "no capabilities means no CapabilityCatalog")
				assert.Equal(t, gemara.ThreatCatalogArtifact.String(), output.Artifacts[0].Type)
				assert.Contains(t, output.Artifacts[0].Content, "v1.0.0")
				assert.NotContains(t, output.Artifacts[0].Content, "imports")
			},
		},
		{
			name: "CapabilityCatalog title fallback appends suffix",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: `metadata:
  id: TEST
  type: ThreatCatalog
  gemara-version: "0.20.0"
  description: Test
  version: 1.0.0
  author:
    id: test
    name: Test
    type: Human
title: My Custom Threats
capabilities:
  - id: TEST.CAP01
    title: Cap One
    description: Test
threats: []
`,
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 2)
				cc := output.Artifacts[1]
				assert.Contains(t, cc.Content, "My Custom Threats - Capabilities",
					"title should fall back to appending ' - Capabilities'")
			},
		},
		{
			name: "ControlCatalog migration",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: testV0ControlCatalog,
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 1)
				assert.Contains(t, output.Message, gemara.ControlCatalogArtifact.String())

				cc := output.Artifacts[0]
				assert.Equal(t, gemara.ControlCatalogArtifact.String(), cc.Type)
				assert.Equal(t, "controls.yaml", cc.SuggestedFilename)
				assert.Contains(t, cc.Content, "v1.0.0")
				assert.Contains(t, cc.Content, "TEST.C01")
				assert.Contains(t, cc.Content, "groups:", "families should be renamed to groups")
				assert.NotContains(t, cc.Content, "families:", "families should not appear in migrated output")
				assert.Contains(t, cc.Content, "group: test-family", "control.family should be renamed to control.group")
				assert.NotContains(t, cc.Content, "family:", "family should not appear in migrated controls")
				assert.Contains(t, cc.Content, "applicability-groups:", "applicability-categories should be renamed to applicability-groups")
				assert.NotContains(t, cc.Content, "applicability-categories:", "applicability-categories should not appear in migrated output")
			},
		},
		{
			name: "ControlCatalog with imported-controls maps to imports",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: testV0ControlCatalogWithImportedControls,
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 1)
				cc := output.Artifacts[0]
				assert.Contains(t, cc.Content, "imports:", "imported-controls should be mapped to imports")
				assert.Contains(t, cc.Content, "EXT.C01", "imported control entries should be preserved")
				assert.Contains(t, cc.Content, "EXT.C02", "imported control entries should be preserved")
				assert.NotContains(t, cc.Content, "imported-controls", "old field name should not appear in output")
			},
		},
		{
			name: "missing type supplied by artifact_type",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: `metadata:
  id: TEST
  gemara-version: "0.20.0"
  description: Test
  version: 1.0.0
  author:
    id: test
    name: Test
    type: Human
title: Test Control Catalog
families:
  - id: test-family
    title: Test Family
    description: Test family
controls:
  - id: TEST.C01
    family: test-family
    title: Test Control
    objective: Test objective
    assessment-requirements:
      - id: TEST.C01.TR01
        text: Test requirement
`,
				ArtifactType: gemara.ControlCatalogArtifact.String(),
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 1)
				assert.Equal(t, gemara.ControlCatalogArtifact.String(), output.Artifacts[0].Type)
				assert.Contains(t, output.Artifacts[0].Content, "v1.0.0")
			},
		},
		{
			name: "missing gemara-version supplied by gemara_version",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: `metadata:
  id: TEST
  type: ControlCatalog
  description: Test
  version: 1.0.0
  author:
    id: test
    name: Test
    type: Human
title: Test Control Catalog
families:
  - id: test-family
    title: Test Family
    description: Test family
controls:
  - id: TEST.C01
    family: test-family
    title: Test Control
    objective: Test objective
    assessment-requirements:
      - id: TEST.C01.TR01
        text: Test requirement
`,
				GemaraVersion: "0.20.0",
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 1)
				assert.Equal(t, gemara.ControlCatalogArtifact.String(), output.Artifacts[0].Type)
				assert.Contains(t, output.Artifacts[0].Content, "v1.0.0")
			},
		},
		{
			name: "missing metadata block constructed from params",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: `title: Test Control Catalog
families:
  - id: test-family
    title: Test Family
    description: Test family
controls:
  - id: TEST.C01
    family: test-family
    title: Test Control
    objective: Test objective
    assessment-requirements:
      - id: TEST.C01.TR01
        text: Test requirement
`,
				ArtifactType:  gemara.ControlCatalogArtifact.String(),
				GemaraVersion: "0.20.0",
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				require.Len(t, output.Artifacts, 1)
				assert.Equal(t, gemara.ControlCatalogArtifact.String(), output.Artifacts[0].Type)
			},
		},
		{
			name: "yaml type takes precedence over artifact_type",
			input: InputMigrateGemaraArtifact{
				ArtifactContent: testV0ControlCatalog,
				ArtifactType:    gemara.ThreatCatalogArtifact.String(),
			},
			validateOutput: func(t *testing.T, output OutputMigrateGemaraArtifact) {
				assert.Equal(t, gemara.ControlCatalogArtifact.String(), output.Artifacts[0].Type,
					"YAML metadata.type should take precedence over input artifact_type")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := MigrateGemaraArtifact(context.Background(), nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}
