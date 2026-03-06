// SPDX-License-Identifier: Apache-2.0

package parsers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gemaraproj/gemara-mcp/internal/evidence"
	"github.com/goccy/go-yaml"
)

// YAMLParser parses YAML and JSON configuration files into EvidenceChunks.
// It flattens the top-level keys of the document, treating each key-value
// pair as a separate chunk with the key as the SectionPath.
type YAMLParser struct{}

func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

func (p *YAMLParser) Name() string {
	return "yaml"
}

func (p *YAMLParser) CanHandle(source evidence.EvidenceSource) bool {
	switch strings.ToLower(source.Format) {
	case "yaml", "yml", "json":
		return true
	}
	content := strings.TrimSpace(string(source.Content))
	// JSON object
	if strings.HasPrefix(content, "{") {
		return true
	}
	// Plain YAML: key: value at the start
	if len(content) > 0 && strings.Contains(strings.SplitN(content, "\n", 2)[0], ":") {
		// Avoid stealing from Dockerfile or Markdown parsers
		if !strings.HasPrefix(content, "#") && !strings.HasPrefix(content, "FROM") {
			return true
		}
	}
	return false
}

func (p *YAMLParser) Parse(_ context.Context, source evidence.EvidenceSource) ([]evidence.EvidenceChunk, error) {
	var doc map[string]interface{}
	if err := yaml.Unmarshal(source.Content, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML/JSON: %w", err)
	}

	var chunks []evidence.EvidenceChunk
	for key, value := range doc {
		rendered, err := yaml.Marshal(value)
		if err != nil {

			rendered = []byte(fmt.Sprintf("%v", value))
		}
		chunks = append(chunks, evidence.EvidenceChunk{
			Text:        fmt.Sprintf("%s: %s", key, strings.TrimSpace(string(rendered))),
			SourceID:    source.ID,
			SectionPath: key,
			Confidence:  0.80,
		})
	}
	return chunks, nil
}
