// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSchema() string {
	return `
		#Person: {
			name: string
			age:  int & >=0
		}
	`
}

func TestValidate(t *testing.T) {
	ctx := cuecontext.New()
	val := ctx.CompileString(testSchema())
	require.NoError(t, val.Err())

	tests := []struct {
		name       string
		definition string
		yaml       string
		wantValid  bool
		wantErr    bool
	}{
		{
			name:       "valid YAML",
			definition: "#Person",
			yaml:       "name: Alice\nage: 30",
			wantValid:  true,
		},
		{
			name:       "wrong field type",
			definition: "#Person",
			yaml:       "name: Alice\nage: not-a-number",
			wantValid:  false,
		},
		{
			name:       "missing required field",
			definition: "#Person",
			yaml:       "name: Alice",
			wantValid:  false,
		},
		{
			name:       "extra field rejected",
			definition: "#Person",
			yaml:       "name: Alice\nage: 30\nemail: a@b.com",
			wantValid:  false,
		},
		{
			name:       "invalid YAML syntax",
			definition: "#Person",
			yaml:       "invalid: yaml: [unclosed",
			wantValid:  false,
		},
		{
			name:       "definition not found",
			definition: "#NonExistent",
			yaml:       "name: Alice\nage: 30",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Validate(val, tt.definition, tt.yaml)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, result.Valid)
			assert.NotEmpty(t, result.Message)
			if tt.wantValid {
				assert.Empty(t, result.Errors)
			} else {
				assert.NotEmpty(t, result.Errors)
			}
		})
	}
}
