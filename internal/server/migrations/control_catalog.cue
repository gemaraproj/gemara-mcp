// SPDX-License-Identifier: Apache-2.0

// ControlCatalog migration
//
// Bumps gemara-version; renames families→groups and control.family→control.group.

package migrate

import gemara "github.com/gemaraproj/gemara@v1"

input: {...}
target_gemara_version: string

output: gemara.#ControlCatalog & {
	metadata: {
		if input.metadata.id != _|_ {id: input.metadata.id}
		if input.metadata.id == _|_ {id: "REPLACE ME"}
		type: "ControlCatalog"
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
		if input.metadata."mapping-references" != _|_ {"mapping-references": input.metadata."mapping-references"}
		if input.metadata."applicability-categories" != _|_ {"applicability-groups": input.metadata."applicability-categories"}
		if input.metadata."applicability-groups" != _|_ {"applicability-groups": input.metadata."applicability-groups"}
	}

	if input.title != _|_ {title: input.title}
	if input.title == _|_ {title: "REPLACE ME"}

	if input.families != _|_ {
		groups: input.families
	}
	if input.groups != _|_ {
		groups: input.groups
	}

	controls: [for c in input.controls {
		if c.family != _|_ {
			group: c.family
		}
		if c.group != _|_ {
			group: c.group
		}
		id:    c.id
		title: c.title
		if c.objective != _|_ {objective: c.objective}
		if c.objective == _|_ {objective: "REPLACE ME"}
		if c."assessment-requirements" != _|_ {
			"assessment-requirements": [for ar in c."assessment-requirements" {
				id:   ar.id
				text: ar.text
				if ar.applicability != _|_ {applicability: ar.applicability}
				if ar.applicability == _|_ {applicability: ["REPLACE ME"]}
				if ar.recommendation != _|_ {recommendation: ar.recommendation}
				if ar.state != _|_ {state: ar.state}
				if ar.state == _|_ {state: "Active"}
			}]
		}
		if c."assessment-requirements" == _|_ {
			"assessment-requirements": [{
				id:            "REPLACE ME"
				text:          "REPLACE ME"
				applicability: ["REPLACE ME"]
			}]
		}
		if c.guidelines != _|_ {guidelines: c.guidelines}
		if c.threats != _|_ {threats: c.threats}
		if c.state != _|_ {state: c.state}
		if c.state == _|_ {state: "Active"}
	}]

  if input."imported-controls" != _|_ {
		imports: [...gemara.#MultiEntryMapping] & input."imported-controls"
	}
	if input.imports != _|_ {
		if input.imports.controls != _|_ {
			imports: [...gemara.#MultiEntryMapping] & input.imports.controls
		}
		if input.imports.controls == _|_ {
			imports: [...gemara.#MultiEntryMapping] & input.imports
		}
	}
}
