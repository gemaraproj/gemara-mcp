// SPDX-License-Identifier: Apache-2.0

package evidence

import "context"

type EvidenceChunk struct {
	Text        string
	SourceID    string
	SectionPath string
	Confidence  float64
}

type SchemaCandidate struct {
	TargetField string  `json:"field"`
	Value       string  `json:"value"`
	SourceRef   string  `json:"source"`
	Confidence  float64 `json:"confidence"`
}

// EvidenceSource describes the raw input to the evidence pipeline.
type EvidenceSource struct {
	// Content is the raw document content.
	Content []byte
	Format  string
	ID      string
}
type EvidenceParser interface {
	CanHandle(source EvidenceSource) bool
	Parse(ctx context.Context, source EvidenceSource) ([]EvidenceChunk, error)
	Name() string
}
