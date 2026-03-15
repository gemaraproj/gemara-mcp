// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatDefinitions(t *testing.T) {
	ctx := cuecontext.New()
	val := ctx.CompileString(`
		// Person describes an individual.
		#Person: {
			name: string
			age:  int
		}
	`)
	require.NoError(t, val.Err())

	formatted, err := FormatDefinitions(val)
	require.NoError(t, err)
	assert.Contains(t, formatted, "#Person")
	assert.Contains(t, formatted, "name")
	assert.Contains(t, formatted, "age")
}

func TestFormatDefinitionsEmptyValue(t *testing.T) {
	_, err := FormatDefinitions(cue.Value{})
	assert.Error(t, err)
}
