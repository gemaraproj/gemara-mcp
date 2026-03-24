// SPDX-License-Identifier: Apache-2.0

package server

import (
	gemara "github.com/gemaraproj/go-gemara"
)

const (
	LexiconResourceURI            = "gemara://lexicon"
	SchemaDocsResourceURI         = "gemara://schema/definitions"
	SchemaDocsResourceURITemplate = "gemara://schema/definitions{?version}"

	// PolicyRiskWizardGemaraVersion is metadata.gemara-version for Policy and Risk Catalog wizards
	// (aligned with Gemara v1.0.0-rc.0).
	PolicyRiskWizardGemaraVersion = "1.0.0-rc.0"
)

// DefaultGemaraVersion is derived from the go-gemara SDK's supported schema version.
var DefaultGemaraVersion = gemara.SchemaVersion
