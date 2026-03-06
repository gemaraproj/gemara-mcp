// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"strings"
)

// fieldRule maps a set of trigger keywords to a target Gemara schema field.
type fieldRule struct {
	keywords    []string
	targetField string
}

// schemaFieldRules defines the keyword-to-field mapping table used by the SchemaMapper.
// Rules are evaluated in order; the first match wins.
var schemaFieldRules = []fieldRule{
	{keywords: []string{"identifier", "id:", "control id", "policy id"}, targetField: "metadata.id"},
	{keywords: []string{"title:", "name:", "policy name", "control name"}, targetField: "metadata.title"},
	{keywords: []string{"version:", "revision:"}, targetField: "metadata.version"},
	{keywords: []string{"objective", "goal", "purpose", "intent"}, targetField: "controls[].objective"},
	{keywords: []string{"control statement", "requirement", "must ", "shall ", "required to"}, targetField: "controls[].statement"},
	{keywords: []string{"assessment", "verify", "verification", "audit", "check"}, targetField: "controls[].assessment"},
	{keywords: []string{"implementation", "procedure", "how to", "steps to"}, targetField: "controls[].implementation"},
	{keywords: []string{"parameter", "setting", "configuration", "config value"}, targetField: "controls[].parameters[]"},
	{keywords: []string{"reference", "see also", "related", "maps to"}, targetField: "metadata.references[]"},
	{keywords: []string{"scope", "applies to", "applicability"}, targetField: "metadata.scope"},
	{keywords: []string{"description", "overview", "summary", "background"}, targetField: "metadata.description"},
}

// SchemaMapper maps a list of EvidenceChunks to SchemaCandidate proposals.
type SchemaMapper struct{}

// NewSchemaMapper creates a new SchemaMapper.
func NewSchemaMapper() *SchemaMapper {
	return &SchemaMapper{}
}

func (m *SchemaMapper) Map(chunks []EvidenceChunk) []SchemaCandidate {
	candidates := make([]SchemaCandidate, 0, len(chunks))
	for _, chunk := range chunks {
		candidate := m.mapChunk(chunk)
		if candidate != nil {
			candidates = append(candidates, *candidate)
		}
	}
	return candidates
}

func (m *SchemaMapper) mapChunk(chunk EvidenceChunk) *SchemaCandidate {
	lower := strings.ToLower(chunk.Text)

	for _, rule := range schemaFieldRules {
		for _, kw := range rule.keywords {
			if strings.Contains(lower, kw) {

				mappingConfidence := 0.75
				combined := mappingConfidence * chunk.Confidence

				return &SchemaCandidate{
					TargetField: rule.targetField,
					Value:       normalizeValue(chunk.Text),
					SourceRef:   chunk.SourceID + " / " + chunk.SectionPath,
					Confidence:  combined,
				}
			}
		}
	}

	return nil
}

func normalizeValue(text string) string {
	lines := strings.Split(text, "\n")
	parts := make([]string, 0, len(lines))
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, " ")
}
