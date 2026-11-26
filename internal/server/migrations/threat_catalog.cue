// SPDX-License-Identifier: Apache-2.0

// ThreatCatalog migration
//
// Bumps gemara-version and removes inline capabilities (extracted
// into a standalone CapabilityCatalog by capability_catalog.cue).
// Adds a mapping-references entry so threat-level capabilities
// references resolve to the extracted CapabilityCatalog.

package migrate

import (
	"list"
	gemara "github.com/gemaraproj/gemara@v1"
)

input: {...}
target_gemara_version:   string
capability_catalog_title: string | *""
capabilities_extracted:   bool | *false

output: gemara.#ThreatCatalog & {
	// The extracted CapabilityCatalog's id (see capability_catalog.cue).
	let _capID = [if input.metadata.id != _|_ {input.metadata.id + "-capabilities"}, "REPLACE ME"][0]
	// migrate.go sets capabilities_extracted when it extracts a
	// CapabilityCatalog, so only then is there a catalog to point at.
	let _extracted = capabilities_extracted

	metadata: {
		if input.metadata.id != _|_ {id: input.metadata.id}
		if input.metadata.id == _|_ {id: "REPLACE ME"}
		type: "ThreatCatalog"
		"gemara-version": target_gemara_version
		if input.metadata.description != _|_ {description: input.metadata.description}
		if input.metadata.description == _|_ {description: "REPLACE ME"}
		if input.metadata.version != _|_ {version: input.metadata.version}
		if input.metadata.author != _|_ {author: input.metadata.author}
		if input.metadata.author == _|_ {
			author: {
				id:   "REPLACE ME"
				name: "REPLACE ME"
				type: "Human"
			}
		}

		let _capRef = [{
			id:          _capID
			title:       capability_catalog_title
			if input.metadata.version != _|_ {version: input.metadata.version}
			if input.metadata.version == _|_ {version: "REPLACE ME"}
			description: "Capabilities extracted from " + input.title
		}]

		if _extracted {
			if input.metadata."mapping-references" != _|_ {
				"mapping-references": list.Concat([input.metadata."mapping-references", _capRef])
			}
			if input.metadata."mapping-references" == _|_ {
				"mapping-references": _capRef
			}
		}

		if !_extracted {
			if input.metadata."mapping-references" != _|_ {
				"mapping-references": input.metadata."mapping-references"
			}
		}
	}

	if input.title != _|_ {title: input.title}
	if input.title == _|_ {title: "REPLACE ME"}

	groups: [{id: "REPLACE ME", title: "REPLACE ME", description: "REPLACE ME"}]

	if input."imported-threats" != _|_ {
		imports: [...gemara.#MultiEntryMapping] & input."imported-threats"
	}
	if input.imports != _|_ {
		if input.imports.threats != _|_ {
			imports: [...gemara.#MultiEntryMapping] & input.imports.threats
		}
		if input.imports.threats == _|_ {
			imports: [...gemara.#MultiEntryMapping] & input.imports
		}
	}

	if input.threats != _|_ {
		if input.threats[0] != _|_ {
			threats: [for t in input.threats {
				id:    t.id
				title: t.title
				if t.description != _|_ {description: t.description}
				if t.description == _|_ {description: "REPLACE ME"}
				group: "REPLACE ME"
				// A v0 threat names its catalog's inline capabilities by the
				// catalog's own id. Those are in the extracted CapabilityCatalog
				// now, so point such mappings at its id. Mappings to any other
				// catalog, or when nothing was extracted, are kept as they are.
				if t.capabilities != _|_ {
					capabilities: [for m in t.capabilities {
						for k, v in m if k != "reference-id" {(k): v}
						if _extracted && input.metadata.id != _|_ {
							if m."reference-id" == input.metadata.id {"reference-id": _capID}
							if m."reference-id" != input.metadata.id {"reference-id": m."reference-id"}
						}
						if !_extracted || input.metadata.id == _|_ {"reference-id": m."reference-id"}
					}]
				}
				if t.capabilities == _|_ {
					capabilities: [{
						"reference-id": "REPLACE ME"
						entries: [{"reference-id": "REPLACE ME"}]
					}]
				}
				if t.vectors != _|_ {vectors: t.vectors}
				if t.actors != _|_ {actors: t.actors}
			}]
		}
	}
}
