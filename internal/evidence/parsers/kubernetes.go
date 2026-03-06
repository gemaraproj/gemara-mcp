// SPDX-License-Identifier: Apache-2.0

package parsers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gemaraproj/gemara-mcp/internal/evidence"
	"github.com/goccy/go-yaml"
)

// kubeManifest is a minimal struct for reading the top-level fields of a
// Kubernetes manifest without pulling in a full k8s client dependency.
type kubeManifest struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec"`
}

// KubernetesParser parses Kubernetes manifests into EvidenceChunks.
// It extracts security-relevant fields (image, securityContext, env, resources)
// from workload specs, making them available for control mapping.
type KubernetesParser struct{}

// NewKubernetesParser creates a new KubernetesParser.
func NewKubernetesParser() *KubernetesParser {
	return &KubernetesParser{}
}

func (p *KubernetesParser) Name() string {
	return "kubernetes"
}

// CanHandle returns true for sources with a "kubernetes" or "k8s" format hint,
// or whose content contains the characteristic apiVersion/kind YAML fields.
func (p *KubernetesParser) CanHandle(source evidence.EvidenceSource) bool {
	switch strings.ToLower(source.Format) {
	case "kubernetes", "k8s":
		return true
	}
	content := string(source.Content)
	return strings.Contains(content, "apiVersion:") && strings.Contains(content, "kind:")
}

// Parse extracts security-relevant fields from a Kubernetes manifest.
// Multi-document YAML (separated by '---') is split and each document parsed independently.
func (p *KubernetesParser) Parse(_ context.Context, source evidence.EvidenceSource) ([]evidence.EvidenceChunk, error) {
	docs := strings.Split(string(source.Content), "\n---")
	var chunks []evidence.EvidenceChunk

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}
		docChunks, err := p.parseDocument([]byte(doc), source.ID, i)
		if err != nil {

			continue
		}
		chunks = append(chunks, docChunks...)
	}
	return chunks, nil
}

func (p *KubernetesParser) parseDocument(content []byte, sourceID string, docIndex int) ([]evidence.EvidenceChunk, error) {
	var manifest kubeManifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	resourceRef := fmt.Sprintf("%s/%s", manifest.Kind, manifest.APIVersion)
	if name, ok := manifest.Metadata["name"]; ok {
		resourceRef = fmt.Sprintf("%s/%v (doc %d)", manifest.Kind, name, docIndex)
	}

	var chunks []evidence.EvidenceChunk

	// Emit a chunk for the resource identity itself
	if manifest.Kind != "" {
		chunks = append(chunks, evidence.EvidenceChunk{
			Text:        fmt.Sprintf("kind: %s\napiVersion: %s", manifest.Kind, manifest.APIVersion),
			SourceID:    sourceID,
			SectionPath: resourceRef + " / identity",
			Confidence:  0.90,
		})
	}

	if manifest.Spec != nil {
		chunks = append(chunks, p.extractSpecChunks(manifest.Spec, sourceID, resourceRef)...)
	}

	return chunks, nil
}

// extractSpecChunks walks the spec looking for security-relevant keys.
func (p *KubernetesParser) extractSpecChunks(spec map[string]interface{}, sourceID, resourceRef string) []evidence.EvidenceChunk {
	securityKeys := []string{
		"securityContext", "containers", "initContainers",
		"volumes", "serviceAccountName", "hostNetwork",
		"hostPID", "hostIPC", "resources", "env", "image",
	}

	var chunks []evidence.EvidenceChunk
	for _, key := range securityKeys {
		val, ok := spec[key]
		if !ok {
			continue
		}
		rendered, err := yaml.Marshal(val)
		if err != nil {
			rendered = []byte(fmt.Sprintf("%v", val))
		}
		chunks = append(chunks, evidence.EvidenceChunk{
			Text:        fmt.Sprintf("%s:\n%s", key, strings.TrimSpace(string(rendered))),
			SourceID:    sourceID,
			SectionPath: resourceRef + " / spec." + key,
			Confidence:  0.88,
		})
	}
	return chunks
}
