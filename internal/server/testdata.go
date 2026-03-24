// SPDX-License-Identifier: Apache-2.0

package server

const testV0ThreatCatalog = `metadata:
  id: TEST
  type: ThreatCatalog
  gemara-version: "0.20.0"
  description: Test threat catalog
  version: 1.0.0
  author:
    id: test
    name: Test Author
    type: Human
  mapping-references:
    - id: REF1
      title: Reference 1
      version: "1.0"
      url: https://example.com
      description: Test reference
title: Test Security Threat Catalog
capabilities:
  - id: TEST.CAP01
    title: Capability One
    description: First capability
  - id: TEST.CAP02
    title: Capability Two
    description: Second capability
threats:
  - id: TEST.THR01
    title: Threat One
    description: First threat
    capabilities:
      - reference-id: TEST
        entries:
          - reference-id: TEST.CAP01
            remarks: Related to cap 1
`

const testV0ThreatCatalogWithImports = `metadata:
  id: TEST
  type: ThreatCatalog
  gemara-version: "0.20.0"
  description: Test threat catalog with imports
  version: 1.0.0
  author:
    id: test
    name: Test Author
    type: Human
  mapping-references:
    - id: EXT
      title: External Reference
      version: "1.0"
      description: External threat source
title: Test Security Threat Catalog
imports:
  - reference-id: EXT
    entries:
      - reference-id: EXT.THR01
        remarks: Imported threat
      - reference-id: EXT.THR02
        remarks: Another imported threat
capabilities:
  - id: TEST.CAP01
    title: Capability One
    description: First capability
threats:
  - id: TEST.THR01
    title: Threat One
    description: First threat
    capabilities:
      - reference-id: TEST
        entries:
          - reference-id: TEST.CAP01
            remarks: Related to cap 1
`

const testV0ThreatCatalogWithImportedThreats = `metadata:
  id: TEST
  type: ThreatCatalog
  gemara-version: "0.20.0"
  description: Test threat catalog with legacy import fields
  version: 1.0.0
  author:
    id: test
    name: Test Author
    type: Human
  mapping-references:
    - id: EXT
      title: External Reference
      version: "1.0"
      url: https://example.com/external
      description: External catalog reference
title: Test Security Threat Catalog
imported-capabilities:
  - reference-id: EXT
    entries:
      - reference-id: EXT.CP01
        remarks: External capability one
      - reference-id: EXT.CP02
        remarks: External capability two
imported-threats:
  - reference-id: EXT
    entries:
      - reference-id: EXT.THR01
        remarks: External threat one
      - reference-id: EXT.THR02
        remarks: External threat two
capabilities:
  - id: TEST.CAP01
    title: Capability One
    description: First capability
threats:
  - id: TEST.THR01
    title: Threat One
    description: First threat
    capabilities:
      - reference-id: TEST
        entries:
          - reference-id: TEST.CAP01
      - reference-id: EXT
        entries:
          - reference-id: EXT.CP03
`

const testV0ControlCatalog = `metadata:
  id: TEST
  type: ControlCatalog
  gemara-version: "0.20.0"
  description: Test control catalog
  version: 1.0.0
  author:
    id: test
    name: Test Author
    type: Human
  applicability-categories:
    - id: default
      title: Default
      description: Default applicability
title: Test Control Catalog
families:
  - id: test-family
    title: Test Family
    description: Test family
controls:
  - id: TEST.C01
    family: test-family
    title: Test Control
    objective: Test objective
    assessment-requirements:
      - id: TEST.C01.TR01
        text: Test requirement
        applicability:
          - default
`

const testV0ControlCatalogWithImportedControls = `metadata:
  id: TEST
  type: ControlCatalog
  gemara-version: "0.20.0"
  description: Test control catalog with legacy import fields
  version: 1.0.0
  author:
    id: test
    name: Test Author
    type: Human
  applicability-categories:
    - id: default
      title: Default
      description: Default applicability
  mapping-references:
    - id: EXT
      title: External Reference
      version: "1.0"
      description: External control catalog reference
title: Test Control Catalog
families:
  - id: test-family
    title: Test Family
    description: Test family
imported-controls:
  - reference-id: EXT
    entries:
      - reference-id: EXT.C01
        remarks: External control one
      - reference-id: EXT.C02
        remarks: External control two
controls:
  - id: TEST.C01
    family: test-family
    title: Test Control
    objective: Test objective
    assessment-requirements:
      - id: TEST.C01.TR01
        text: Test requirement
        applicability:
          - default
`
