// SPDX-License-Identifier: Apache-2.0

// CapabilityCatalog extraction
//
// Extracts a standalone CapabilityCatalog from a ThreatCatalog's
// inline capabilities. The catalog gets its own id, the threat
// catalog's id with "-capabilities" appended, so the two artifacts
// don't share one. threat_catalog.cue points the threats' capability
// references at that id.

package migrate

import gemara "github.com/gemaraproj/gemara@v1"

input: {...}
target_gemara_version: string
capability_title:      string

output: gemara.#CapabilityCatalog & {
	metadata: {
		if input.metadata.id != _|_ {id: input.metadata.id + "-capabilities"}
		if input.metadata.id == _|_ {id: "REPLACE ME"}
		type: "CapabilityCatalog"
		"gemara-version": target_gemara_version
		if input.title != _|_ {description: "Capabilities extracted from " + input.title}
		if input.title == _|_ {description: "REPLACE ME"}
		if input.metadata.version != _|_ {version: input.metadata.version}
		if input.metadata.author != _|_ {author: input.metadata.author}
		if input.metadata.author == _|_ {
			author: {
				id:   "REPLACE ME"
				name: "REPLACE ME"
				type: "Human"
			}
		}
		if input.metadata."mapping-references" != _|_ {"mapping-references": input.metadata."mapping-references"}
	}

	title: capability_title

	groups: [{id: "REPLACE ME", title: "REPLACE ME", description: "REPLACE ME"}]

	capabilities: [for c in input.capabilities {
		id:    c.id
		title: c.title
		if c.description != _|_ {description: c.description}
		if c.description == _|_ {description: "REPLACE ME"}
		group: "REPLACE ME"
	}]

	if input."imported-capabilities" != _|_ {
		imports: [...gemara.#MultiEntryMapping] & input."imported-capabilities"
	}
	if input.imports != _|_ {
		if input.imports.capabilities != _|_ {
			imports: [...gemara.#MultiEntryMapping] & input.imports.capabilities
		}
	}
}
