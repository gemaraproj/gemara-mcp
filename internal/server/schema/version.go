// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"context"
	"fmt"
	"log/slog"

	"cuelang.org/go/mod/modconfig"
)

// CUEVersionResolver queries the CUE registry for all versions of a module
// and returns the highest semver tag.
type CUEVersionResolver struct {
	modulePath string
}

// NewCUEVersionResolver creates a CUEVersionResolver for the given module path.
func NewCUEVersionResolver(modulePath string) *CUEVersionResolver {
	return &CUEVersionResolver{modulePath: modulePath}
}

func (r *CUEVersionResolver) Fetch(ctx context.Context) (string, string, error) {
	reg, err := modconfig.NewRegistry(nil)
	if err != nil {
		return "", "", fmt.Errorf("creating CUE registry: %w", err)
	}

	// modregistry.Client.ModuleVersions returns results sorted in semver order.
	versions, err := reg.ModuleVersions(ctx, r.modulePath)
	if err != nil {
		return "", "", fmt.Errorf("listing module versions for %s: %w", r.modulePath, err)
	}

	if len(versions) == 0 {
		return "", "", fmt.Errorf("no versions found for module %s", r.modulePath)
	}

	latest := versions[len(versions)-1]
	slog.Info("resolved latest module version", "module", r.modulePath, "version", latest)
	return latest, r.modulePath, nil
}
