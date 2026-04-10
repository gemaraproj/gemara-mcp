// SPDX-License-Identifier: Apache-2.0

package server

import (
	gemara "github.com/gemaraproj/go-gemara"
)

const (
	LexiconResourceURI            = "gemara://lexicon"
	SchemaDocsResourceURI         = "gemara://schema/definitions"
	SchemaDocsResourceURITemplate = "gemara://schema/definitions{?version}"
)

// DefaultGemaraVersion is derived from the go-gemara SDK's supported schema version.
var DefaultGemaraVersion = gemara.SchemaVersion
