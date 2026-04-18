// SPDX-License-Identifier: Apache-2.0

package httpserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testIssuer = "https://auth.example.com/"

func newTestHandler(t *testing.T, cfg *Config) http.Handler {
	t.Helper()
	stub := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return buildHandler(stub, cfg)
}

func TestBuildHandler(t *testing.T) {
	stub := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("no auth flags — unprotected", func(t *testing.T) {
		cfg := Config{Address: "127.0.0.1:8080"}
		handler := buildHandler(stub, &cfg)
		require.NotNil(t, handler)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, metadataEndpointPath, nil))
		assert.Equal(t, http.StatusNotFound, rec.Code, "metadata endpoint should not be registered")
	})

	t.Run("gateway mode — metadata only", func(t *testing.T) {
		cfg := Config{
			Address:       "127.0.0.1:8080",
			AuthServerURL: testIssuer,
		}
		handler := buildHandler(stub, &cfg)
		require.NotNil(t, handler)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, metadataEndpointPath, nil))
		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.Equal(t, "http://127.0.0.1:8080/mcp", body["resource"])
		servers, ok := body["authorization_servers"].([]any)
		require.True(t, ok)
		assert.Equal(t, testIssuer, servers[0])
	})

	t.Run("gateway mode — metadata includes scopes_supported", func(t *testing.T) {
		cfg := Config{
			Address:        "127.0.0.1:8080",
			AuthServerURL:  testIssuer,
			RequiredScopes: []string{"gemara:read", "gemara:write"},
		}
		handler := buildHandler(stub, &cfg)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, metadataEndpointPath, nil))
		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		scopes, ok := body["scopes_supported"].([]any)
		require.True(t, ok, "metadata must include scopes_supported")
		assert.Equal(t, []any{"gemara:read", "gemara:write"}, scopes)
	})

	t.Run("gateway mode — MCP requests pass without token", func(t *testing.T) {
		cfg := Config{
			Address:       "127.0.0.1:8080",
			AuthServerURL: testIssuer,
		}
		handler := buildHandler(stub, &cfg)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, mcpEndpointPath, strings.NewReader("{}")))
		assert.Equal(t, http.StatusOK, rec.Code, "gateway mode delegates auth to upstream proxy")
	})
}

func TestHealthEndpoint(t *testing.T) {
	handler := newTestHandler(t, &Config{Address: "127.0.0.1:8080"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, healthEndpointPath, nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "healthy", body["status"])
	assert.NotContains(t, body, "version", "health endpoint must not expose server version")
}

func TestRootEndpoint(t *testing.T) {
	handler := newTestHandler(t, &Config{Address: "127.0.0.1:8080"})

	t.Run("returns endpoint listing without version", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.Equal(t, ServerName, body["name"])
		assert.Contains(t, body, "endpoints")
		assert.NotContains(t, body, "version", "root endpoint must not expose server version")
	})

	t.Run("metadata endpoint listed when auth configured", func(t *testing.T) {
		authHandler := newTestHandler(t, &Config{
			Address:       "127.0.0.1:8080",
			AuthServerURL: "https://auth.example.com/",
		})
		rec := httptest.NewRecorder()
		authHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		endpoints, ok := body["endpoints"].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, endpoints, metadataEndpointPath)
	})

	t.Run("metadata endpoint not listed without auth", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		var body map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		endpoints, ok := body["endpoints"].(map[string]any)
		require.True(t, ok)
		assert.NotContains(t, endpoints, metadataEndpointPath)
	})

	t.Run("unknown path returns 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nonexistent", nil))

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestLimitBody(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := limitBody(inner)

	t.Run("body within limit succeeds", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("ok")))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("body exceeding limit is rejected", func(t *testing.T) {
		oversized := strings.NewReader(strings.Repeat("x", maxRequestBodyBytes+1))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", oversized))
		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	})
}
