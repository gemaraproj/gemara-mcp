// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/gemaraproj/go-gemara"
	"github.com/goccy/go-yaml"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed migrations/threat_catalog.cue
var threatMigrationCUE string

//go:embed migrations/capability_catalog.cue
var capabilityMigrationCUE string

//go:embed migrations/control_catalog.cue
var controlMigrationCUE string

//go:embed migrations/cue.mod/module.cue
var moduleCUE string

const migrateOverlayDir = "/cue/migrate"


// MetadataMigrateGemaraArtifact describes the MigrateGemaraArtifact tool.
var MetadataMigrateGemaraArtifact = &mcp.Tool{
	Name:        "migrate_gemara_artifact",
	Description: "Migrate a Gemara artifact to v1 schema using CUE transformations. When the artifact is missing metadata fields (common in older v0 artifacts), use artifact_type and gemara_version to supply them.",
	InputSchema: map[string]interface{}{
		"type":     "object",
		"required": []string{"artifact_content"},
		"properties": map[string]interface{}{
			"artifact_content": map[string]interface{}{
				"type":        "string",
				"description": "YAML content of the Gemara artifact to migrate",
			},
			"artifact_type": map[string]interface{}{
				"type":        "string",
				"description": "Artifact type when metadata.type is missing. Infer from structure: threats → ThreatCatalog, controls → ControlCatalog.",
				"enum":        []string{gemara.ThreatCatalogArtifact.String(), gemara.ControlCatalogArtifact.String()},
			},
			"gemara_version": map[string]interface{}{
				"type":        "string",
				"description": "Source gemara-version when metadata.gemara-version is missing (e.g. \"0.20.0\").",
			},
		},
	},
}

// InputMigrateGemaraArtifact is the input for the MigrateGemaraArtifact tool.
type InputMigrateGemaraArtifact struct {
	ArtifactContent string `json:"artifact_content"`
	ArtifactType    string `json:"artifact_type"`
	GemaraVersion   string `json:"gemara_version"`
}

// MigratedArtifact represents a single output artifact from the migration.
type MigratedArtifact struct {
	Type              string `json:"type"`
	SuggestedFilename string `json:"suggested_filename"`
	Content           string `json:"content"`
}

// OutputMigrateGemaraArtifact is the output for the MigrateGemaraArtifact tool.
type OutputMigrateGemaraArtifact struct {
	Artifacts []MigratedArtifact `json:"artifacts,omitempty"`
	Changes   []string           `json:"changes"`
	Message   string             `json:"message"`
}

// MigrateGemaraArtifact migrates a Gemara artifact to v1 schema using the pattern - YAML → CUE transformation → YAML.
func MigrateGemaraArtifact(_ context.Context, _ *mcp.CallToolRequest, input InputMigrateGemaraArtifact) (*mcp.CallToolResult, OutputMigrateGemaraArtifact, error) {
	if input.ArtifactContent == "" {
		return nil, OutputMigrateGemaraArtifact{}, fmt.Errorf("artifact_content is required")
	}

	root, err := parseYAMLMap(input.ArtifactContent)
	if err != nil {
		return nil, OutputMigrateGemaraArtifact{}, err
	}

	metaRaw, err := enrichMetadata(root, input)
	if err != nil {
		return nil, OutputMigrateGemaraArtifact{}, err
	}

	metaYAML, err := yaml.Marshal(metaRaw)
	if err != nil {
		return nil, OutputMigrateGemaraArtifact{}, fmt.Errorf("encoding enriched artifact: %w", err)
	}

	var meta gemara.Metadata
	if err := yaml.Unmarshal(metaYAML, &meta); err != nil {
		return nil, OutputMigrateGemaraArtifact{}, fmt.Errorf("invalid artifact metadata: %w", err)
	}

	if meta.GemaraVersion == DefaultGemaraVersion {
		return nil, OutputMigrateGemaraArtifact{}, fmt.Errorf(
			"artifact is already at target gemara-version %q", DefaultGemaraVersion)
	}

	slog.Info("migrating artifact",
		"type", meta.Type,
		"source_version", meta.GemaraVersion,
	)

	cueCtx := cuecontext.New()

	switch meta.Type {
	case gemara.ThreatCatalogArtifact:
		return migrateThreatCatalog(cueCtx, root, meta.GemaraVersion)
	case gemara.ControlCatalogArtifact:
		return migrateSimpleArtifact(cueCtx, root, controlMigrationCUE, meta.Type, "controls.yaml", meta.GemaraVersion)
	default:
		return nil, OutputMigrateGemaraArtifact{}, fmt.Errorf(
			"unsupported artifact type %q for migration", meta.Type)
	}
}

// enrichMetadata fills missing metadata fields from input parameters.
func enrichMetadata(root map[string]interface{}, input InputMigrateGemaraArtifact) (map[string]interface{}, error) {
	metaRaw, ok := root["metadata"].(map[string]interface{})
	if !ok {
		if input.ArtifactType == "" || input.GemaraVersion == "" {
			return metaRaw, fmt.Errorf("artifact missing 'metadata' block; provide artifact_type and gemara_version parameters")
		}
		metaRaw = make(map[string]interface{})
		root["metadata"] = metaRaw
	}

	// ArtifactType is an iota enum; zero value is a valid type,
	// so a missing field would silently unmarshal as that type.
	if _, hasType := metaRaw["type"]; !hasType {
		if input.ArtifactType == "" {
			return metaRaw, fmt.Errorf("artifact missing 'metadata.type'; provide artifact_type parameter")
		}
		metaRaw["type"] = input.ArtifactType
	}

	if _, hasVersion := metaRaw["gemara-version"]; !hasVersion {
		if input.GemaraVersion == "" {
			return metaRaw, fmt.Errorf("artifact missing 'metadata.gemara-version'; provide gemara_version parameter")
		}
		metaRaw["gemara-version"] = input.GemaraVersion
	}

	return metaRaw, nil
}

func migrateThreatCatalog(cueCtx *cue.Context, root map[string]interface{}, sourceVersion string) (*mcp.CallToolResult, OutputMigrateGemaraArtifact, error) {
	tcType := gemara.ThreatCatalogArtifact

	title, _ := root["title"].(string)
	capTitle := strings.Replace(title, "Threat Catalog", "Capability Catalog", 1)
	if capTitle == title {
		capTitle = title + " - Capabilities"
	}

	tcExtras := map[string]interface{}{
		"target_gemara_version":   DefaultGemaraVersion,
		"capability_catalog_title": capTitle,
	}
	tcYAML, err := cueMigrate(cueCtx, threatMigrationCUE, root, tcExtras)
	if err != nil {
		return nil, OutputMigrateGemaraArtifact{}, fmt.Errorf("migrating %s: %w", tcType, err)
	}

	artifacts := []MigratedArtifact{{
		Type:              tcType.String(),
		SuggestedFilename: "threats.yaml",
		Content:           tcYAML,
	}}
	changes := []string{
		fmt.Sprintf("Updated metadata.gemara-version from %q to %q", sourceVersion, DefaultGemaraVersion),
	}

	if capsRaw, ok := root["capabilities"]; ok {
		if caps, ok := capsRaw.([]interface{}); ok && len(caps) > 0 {
			capExtras := map[string]interface{}{
				"target_gemara_version": DefaultGemaraVersion,
				"capability_title":      capTitle,
			}
			capYAML, err := cueMigrate(cueCtx, capabilityMigrationCUE, root, capExtras)
			if err != nil {
				return nil, OutputMigrateGemaraArtifact{}, fmt.Errorf("extracting %s: %w", gemara.CapabilityCatalogArtifact, err)
			}

			capType := gemara.CapabilityCatalogArtifact
			artifacts = append(artifacts, MigratedArtifact{
				Type:              capType.String(),
				SuggestedFilename: "capabilities.yaml",
				Content:           capYAML,
			})
			changes = append(changes,
				fmt.Sprintf("Extracted %d capabilities into standalone %s", len(caps), capType),
				fmt.Sprintf("Removed inline 'capabilities' from %s", tcType),
				fmt.Sprintf("Added mapping-references entry for extracted %s", capType),
			)
		}
	}

	msg := fmt.Sprintf("Migrated %s from %s → %s: %d artifact(s) produced",
		tcType, sourceVersion, DefaultGemaraVersion, len(artifacts))
	slog.Info("migration complete",
		"type", tcType,
		"artifacts", len(artifacts),
		"changes", len(changes),
	)

	return nil, OutputMigrateGemaraArtifact{
		Artifacts: artifacts,
		Changes:   changes,
		Message:   msg,
	}, nil
}

func migrateSimpleArtifact(cueCtx *cue.Context, root map[string]interface{}, cueSrc string, artifactType gemara.ArtifactType, filename, sourceVersion string) (*mcp.CallToolResult, OutputMigrateGemaraArtifact, error) {
	extras := map[string]interface{}{
		"target_gemara_version": DefaultGemaraVersion,
	}
	outputYAML, err := cueMigrate(cueCtx, cueSrc, root, extras)
	if err != nil {
		return nil, OutputMigrateGemaraArtifact{}, fmt.Errorf("migrating %s: %w", artifactType, err)
	}

	changes := []string{
		fmt.Sprintf("Updated metadata.gemara-version from %q to %q", sourceVersion, DefaultGemaraVersion),
	}
	msg := fmt.Sprintf("Migrated %s from %s → %s: 1 artifact(s) produced",
		artifactType, sourceVersion, DefaultGemaraVersion)
	slog.Info("migration complete",
		"type", artifactType,
		"artifacts", 1,
		"changes", len(changes),
	)

	return nil, OutputMigrateGemaraArtifact{
		Artifacts: []MigratedArtifact{{
			Type:              artifactType.String(),
			SuggestedFilename: filename,
			Content:           outputYAML,
		}},
		Changes: changes,
		Message: msg,
	}, nil
}

// cueMigrate loads a CUE migration via the module system (resolving imports),
// fills the input path with YAML data, extracts the output path, and encodes
// the result as YAML.
func cueMigrate(cueCtx *cue.Context, cueSrc string, inputData map[string]interface{}, extras map[string]interface{}) (string, error) {
	// FIXME(jpower432): Is this the correct way to load a local module?
	overlay := map[string]load.Source{
		filepath.Join(migrateOverlayDir, "migrate.cue"):           load.FromString(cueSrc),
		filepath.Join(migrateOverlayDir, "cue.mod", "module.cue"): load.FromString(moduleCUE),
	}

	instances := load.Instances([]string{"."}, &load.Config{
		Dir:     migrateOverlayDir,
		Overlay: overlay,
		Package: "migrate",
	})
	if len(instances) == 0 {
		return "", fmt.Errorf("loading migration CUE: no instances returned")
	}
	if err := instances[0].Err; err != nil {
		return "", fmt.Errorf("loading migration CUE: %w", err)
	}

	migration := cueCtx.BuildInstance(instances[0])
	if err := migration.Err(); err != nil {
		return "", fmt.Errorf("building migration: %w", err)
	}

	unified := migration.FillPath(cue.ParsePath("input"), inputData)
	for k, v := range extras {
		unified = unified.FillPath(cue.ParsePath(k), v)
	}

	outputVal := unified.LookupPath(cue.ParsePath("output"))
	if err := outputVal.Err(); err != nil {
		return "", fmt.Errorf("evaluating migration output: %w", err)
	}

	return cueValueToYAML(outputVal)
}

// cueValueToYAML walks a CUE value tree and marshals it to YAML,
// preserving the field ordering defined by the CUE source.
func cueValueToYAML(v cue.Value) (string, error) {
	ordered, err := cueToOrdered(v)
	if err != nil {
		return "", fmt.Errorf("walking CUE value: %w", err)
	}
	data, err := yaml.Marshal(ordered)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// cueToOrdered recursively converts a CUE value to yaml.MapSlice-based
// types that preserve field ordering during YAML serialization.
func cueToOrdered(v cue.Value) (interface{}, error) {
	switch v.IncompleteKind() {
	case cue.StructKind:
		iter, err := v.Fields()
		if err != nil {
			return nil, err
		}
		var slice yaml.MapSlice
		for iter.Next() {
			child, err := cueToOrdered(iter.Value())
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", iter.Selector().Unquoted(), err)
			}
			slice = append(slice, yaml.MapItem{
				Key:   iter.Selector().Unquoted(),
				Value: child,
			})
		}
		return slice, nil

	case cue.ListKind:
		iter, err := v.List()
		if err != nil {
			return nil, err
		}
		var list []interface{}
		for iter.Next() {
			child, err := cueToOrdered(iter.Value())
			if err != nil {
				return nil, err
			}
			list = append(list, child)
		}
		return list, nil

	case cue.StringKind:
		return v.String()

	case cue.IntKind:
		return v.Int64()

	case cue.FloatKind:
		return v.Float64()

	case cue.BoolKind:
		return v.Bool()

	case cue.NullKind:
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported CUE kind: %v", v.IncompleteKind())
	}
}

// parseYAMLMap unmarshals YAML content into a map for initial inspection.
func parseYAMLMap(content string) (map[string]interface{}, error) {
	var root map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}
	if len(root) == 0 {
		return nil, fmt.Errorf("empty or invalid YAML document")
	}
	return root, nil
}
