// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
	"github.com/gemaraproj/gemara-mcp/internal/tool/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCUESchema = `
	#Person: {
		name: string
		age:  int
	}
`

type staticCUEFetcher struct {
	val    cue.Value
	source string
}

func (f *staticCUEFetcher) Fetch(_ context.Context) (cue.Value, string, error) {
	return f.val, f.source, nil
}

func newTestSchemaCachedFetcher(t *testing.T, source string) *fetcher.CachedFetcher[cue.Value] {
	t.Helper()
	cueCtx := cuecontext.New()
	val := cueCtx.CompileString(testCUESchema)
	require.NoError(t, val.Err())

	cache := fetcher.NewCache[cue.Value](1 * time.Hour)
	f := &staticCUEFetcher{val: val, source: source}
	return fetcher.NewCachedFetcher[cue.Value](f, cache, source)
}

func TestGetSchemaDocsFromCache(t *testing.T) {
	source := gemaraModuleBase + defaultSchemaVersion
	cf := newTestSchemaCachedFetcher(t, source)

	_, output, err := GetSchemaDocs(context.Background(), nil, InputGetSchemaDocs{}, cf)
	require.NoError(t, err)
	assert.Contains(t, output.Documentation, "#Person")
	assert.Contains(t, output.Documentation, "name")
	assert.Equal(t, source, output.URL)
}

func TestGetSchemaDocsVersionedSource(t *testing.T) {
	source := gemaraModuleBase + "v0.19.1"
	cf := newTestSchemaCachedFetcher(t, source)

	_, output, err := GetSchemaDocs(context.Background(), nil, InputGetSchemaDocs{Version: "v0.19.1"}, cf)
	require.NoError(t, err)
	assert.Contains(t, output.Documentation, "#Person")
	assert.Equal(t, source, output.URL)
}

func TestGetSchemaDocsRefreshBypassesCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cache := fetcher.NewCache[cue.Value](1 * time.Hour)
	modulePath := gemaraModuleBase + defaultSchemaVersion
	f := schema.NewCUERegistryFetcher(modulePath)
	cf := fetcher.NewCachedFetcher[cue.Value](f, cache, modulePath)

	_, output, err := GetSchemaDocs(context.Background(), nil, InputGetSchemaDocs{Refresh: true}, cf)
	require.NoError(t, err)
	assert.NotEmpty(t, output.Documentation)
}

func TestGetSchemaDocsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cache := fetcher.NewCache[cue.Value](1 * time.Hour)
	modulePath := gemaraModuleBase + defaultSchemaVersion
	f := schema.NewCUERegistryFetcher(modulePath)
	cf := fetcher.NewCachedFetcher[cue.Value](f, cache, modulePath)

	_, output, err := GetSchemaDocs(context.Background(), nil, InputGetSchemaDocs{}, cf)
	require.NoError(t, err)
	assert.NotEmpty(t, output.Documentation)
	assert.Contains(t, output.URL, "github.com/gemaraproj/gemara@")
}
