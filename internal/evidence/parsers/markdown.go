// SPDX-License-Identifier: Apache-2.0

package parsers

import (
	"context"
	"strings"

	"github.com/gemaraproj/gemara-mcp/internal/evidence"
)

// MarkdownParser parses Markdown governance documents into EvidenceChunks.
// It splits the document on headings (lines starting with '#') and treats
// each section as a separate chunk, using the heading text as the SectionPath.
type MarkdownParser struct{}

// NewMarkdownParser creates a new MarkdownParser.
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

func (p *MarkdownParser) Name() string {
	return "markdown"
}

// CanHandle returns true for sources that use the "markdown" format hint,
// or whose content begins with a Markdown heading or common Markdown patterns.
func (p *MarkdownParser) CanHandle(source evidence.EvidenceSource) bool {
	if strings.EqualFold(source.Format, "markdown") || strings.EqualFold(source.Format, "md") {
		return true
	}
	content := strings.TrimSpace(string(source.Content))
	return strings.HasPrefix(content, "#") || strings.Contains(content, "\n#")
}

func (p *MarkdownParser) Parse(_ context.Context, source evidence.EvidenceSource) ([]evidence.EvidenceChunk, error) {
	lines := strings.Split(string(source.Content), "\n")

	var chunks []evidence.EvidenceChunk
	var currentHeading string
	var currentLines []string

	flush := func() {
		text := strings.TrimSpace(strings.Join(currentLines, "\n"))
		if text == "" {
			return
		}
		sectionPath := currentHeading
		if sectionPath == "" {
			sectionPath = "preamble"
		}
		chunks = append(chunks, evidence.EvidenceChunk{
			Text:        text,
			SourceID:    source.ID,
			SectionPath: sectionPath,
			Confidence:  0.85,
		})
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			// Flush previous section
			flush()
			currentHeading = strings.TrimSpace(strings.TrimLeft(line, "#"))
			currentLines = nil
		} else {
			currentLines = append(currentLines, line)
		}
	}
	flush()

	return chunks, nil
}
