// SPDX-License-Identifier: Apache-2.0

package server

import (
	gemara "github.com/gemaraproj/go-gemara"
)

const (
	LexiconResourceURI            = "gemara://lexicon"
	AboutResourceURI              = "gemara://about"
	SchemaDocsResourceURI         = "gemara://schema/definitions"
	SchemaDocsResourceURITemplate = "gemara://schema/definitions{?version}"

	// aboutURL is the upstream LLM-optimized guidance served by gemara://about.
	// When unreachable (e.g., before this file lands on the website, or offline),
	// the resource falls back to the embedded copy.
	aboutURL = "https://gemara.openssf.org/llms-full.txt"
)

// DefaultGemaraVersion is derived from the go-gemara SDK's supported schema version.
var DefaultGemaraVersion = gemara.SchemaVersion

func boolPtr(b bool) *bool { return &b }
