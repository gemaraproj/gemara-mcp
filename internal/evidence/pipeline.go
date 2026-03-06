// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"context"
	"fmt"
)

type Pipeline struct {
	parsers []EvidenceParser
	mapper  *SchemaMapper
}

// NewPipeline creates a new Pipeline with the provided parsers.
// The SchemaMapper is created internally.
func NewPipeline(parsers ...EvidenceParser) *Pipeline {
	return &Pipeline{
		parsers: parsers,
		mapper:  NewSchemaMapper(),
	}
}

// RunResult is the output of a successful pipeline run.
type RunResult struct {
	Candidates []SchemaCandidate
	ParserUsed string
	ChunkCount int
}

func (p *Pipeline) Run(ctx context.Context, source EvidenceSource) ([]SchemaCandidate, error) {
	result, err := p.RunWithMeta(ctx, source)
	if err != nil {
		return nil, err
	}
	return result.Candidates, nil
}

func (p *Pipeline) RunWithMeta(ctx context.Context, source EvidenceSource) (RunResult, error) {
	parser, err := p.selectParser(source)
	if err != nil {
		return RunResult{}, err
	}

	chunks, err := parser.Parse(ctx, source)
	if err != nil {
		return RunResult{}, fmt.Errorf("parser %q failed: %w", parser.Name(), err)
	}

	candidates := p.mapper.Map(chunks)
	return RunResult{
		Candidates: candidates,
		ParserUsed: parser.Name(),
		ChunkCount: len(chunks),
	}, nil
}

// selectParser returns the first registered parser that can handle the given source.
func (p *Pipeline) selectParser(source EvidenceSource) (EvidenceParser, error) {
	for _, parser := range p.parsers {
		if parser.CanHandle(source) {
			return parser, nil
		}
	}
	return nil, fmt.Errorf("unsupported evidence format: no parser found for source %q (format hint: %q)", source.ID, source.Format)
}

// RegisteredParsers returns the names of all currently registered parsers.
func (p *Pipeline) RegisteredParsers() []string {
	names := make([]string, len(p.parsers))
	for i, parser := range p.parsers {
		names[i] = parser.Name()
	}
	return names
}
