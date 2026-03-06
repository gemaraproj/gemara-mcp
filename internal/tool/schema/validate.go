// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/yaml"
)

// ValidateResult holds the outcome of validating YAML against a CUE definition.
type ValidateResult struct {
	Valid   bool
	Errors  []string
	Message string
}

// Validate checks YAML content against a named CUE definition within the schema.
func Validate(schema cue.Value, definition, yamlContent string) (ValidateResult, error) {
	entrypoint := schema.LookupPath(cue.ParsePath(definition))
	if !entrypoint.Exists() {
		return ValidateResult{}, fmt.Errorf("definition %s not found in schema", definition)
	}

	yamlFile, err := yaml.Extract("artifact.yaml", yamlContent)
	if err != nil {
		return ValidateResult{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("Failed to parse YAML: %v", err)},
			Message: fmt.Sprintf("Validation failed: invalid YAML: %v", err),
		}, nil
	}

	cueCtx := cuecontext.New()
	data := cueCtx.BuildFile(yamlFile)
	if err := data.Err(); err != nil {
		return ValidateResult{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("Failed to build data instance: %v", err)},
			Message: fmt.Sprintf("Validation failed: %v", err),
		}, nil
	}

	unified := entrypoint.Unify(data)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		errorLines := strings.Split(strings.TrimSpace(err.Error()), "\n")
		var errors []string
		for _, line := range errorLines {
			if strings.TrimSpace(line) != "" {
				errors = append(errors, line)
			}
		}
		return ValidateResult{
			Valid:   false,
			Errors:  errors,
			Message: fmt.Sprintf("Validation failed: %v", err),
		}, nil
	}

	return ValidateResult{
		Valid:   true,
		Errors:  []string{},
		Message: "Artifact is valid",
	}, nil
}
