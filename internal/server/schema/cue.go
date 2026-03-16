// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"context"
	"fmt"
	"log/slog"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/mod/modconfig"
)

// CUERegistryFetcher loads a CUE module from the registry and returns a built cue.Value.
type CUERegistryFetcher struct {
	modulePath string
}

// NewCUERegistryFetcher creates a fetcher for the given CUE module path.
func NewCUERegistryFetcher(modulePath string) *CUERegistryFetcher {
	return &CUERegistryFetcher{modulePath: modulePath}
}

// Fetch loads and builds the CUE module from the registry.
// The context parameter is unused because the CUE load API does not
// support context-based cancellation or timeouts.
func (f *CUERegistryFetcher) Fetch(_ context.Context) (cue.Value, string, error) {
	slog.Info("loading schema from registry", "module", f.modulePath)

	reg, err := modconfig.NewRegistry(nil)
	if err != nil {
		return cue.Value{}, "", fmt.Errorf("creating CUE registry: %w", err)
	}

	instances := load.Instances([]string{f.modulePath}, &load.Config{
		Registry: reg,
	})
	if len(instances) == 0 {
		return cue.Value{}, "", fmt.Errorf("loading module %s: no instances returned", f.modulePath)
	}
	if err := instances[0].Err; err != nil {
		return cue.Value{}, "", fmt.Errorf("loading module %s: %w", f.modulePath, err)
	}

	cueCtx := cuecontext.New()
	val := cueCtx.BuildInstance(instances[0])
	if err := val.Err(); err != nil {
		return cue.Value{}, "", fmt.Errorf("building schema: %w", err)
	}

	return val, f.modulePath, nil
}
