// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
)

// FormatDefinitions returns the formatted CUE definitions from a built schema value.
func FormatDefinitions(val cue.Value) (string, error) {
	syn := val.Syntax(
		cue.Definitions(true),
		cue.Optional(true),
		cue.Attributes(true),
		cue.Docs(true),
	)

	formatted, err := format.Node(syn)
	if err != nil {
		return "", fmt.Errorf("formatting schema definitions: %w", err)
	}

	return string(formatted), nil
}
