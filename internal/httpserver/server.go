// SPDX-License-Identifier: Apache-2.0

package httpserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

// Gateway auth mode: serves RFC 9728 metadata so clients know where to
// obtain tokens. An upstream gateway or reverse proxy (Envoy, Istio,
// oauth2-proxy) validates tokens before they reach this server.

// ServerName is the canonical name for this MCP server.
const ServerName = "gemara-mcp"

const (
	mcpEndpointPath      = "/mcp"
	healthEndpointPath   = "/health"
	metadataEndpointPath = "/.well-known/oauth-protected-resource"
	readHeaderTimeout    = 10 * time.Second
	shutdownGracePeriod  = 10 * time.Second
	maxRequestBodyBytes  = 4 << 20 // 4 MiB
)

// Run starts a Streamable HTTP server for the given MCP server.
func Run(ctx context.Context, mcpServer *mcp.Server, cfg *Config) error {
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return mcpServer },
		&mcp.StreamableHTTPOptions{
			Stateless: true,
			Logger:    slog.Default(),
		},
	)

	handler := buildHandler(mcpHandler, cfg)

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	hasTLS := cfg.TLSCert != "" && cfg.TLSKey != ""
	if hasTLS {
		srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS13}
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
		defer cancel()
		slog.Info("shutting down HTTP server")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}
	}()

	slog.Info("HTTP server listening",
		"address", cfg.Address,
		"tls", hasTLS,
		"endpoints", []string{"/", mcpEndpointPath, healthEndpointPath},
	)

	var listenErr error
	if hasTLS {
		listenErr = srv.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
	} else {
		listenErr = srv.ListenAndServe()
	}

	if errors.Is(listenErr, http.ErrServerClosed) {
		return nil
	}
	return listenErr
}

// buildHandler assembles the HTTP handler with operational endpoints
// and optional RFC 9728 metadata.
func buildHandler(mcpHandler http.Handler, cfg *Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler(cfg.AuthServerURL != ""))
	mux.HandleFunc(healthEndpointPath, healthHandler())

	if cfg.AuthServerURL != "" {
		slog.Info("serving OAuth Protected Resource Metadata",
			"auth_server", cfg.AuthServerURL,
			"metadata_path", metadataEndpointPath,
		)
		mux.Handle(metadataEndpointPath,
			auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
				Resource:               ResourceURI(cfg),
				AuthorizationServers:   []string{cfg.AuthServerURL},
				BearerMethodsSupported: []string{"header"},
				ScopesSupported:        cfg.RequiredScopes,
			}),
		)
	}

	mux.Handle(mcpEndpointPath, mcpHandler)

	return limitBody(mux)
}

// rootHandler returns endpoint discovery metadata.
func rootHandler(hasMetadata bool) http.HandlerFunc {
	endpoints := map[string]string{
		"/":                "endpoint listing (this response)",
		mcpEndpointPath:    "MCP Streamable HTTP endpoint",
		healthEndpointPath: "liveness probe",
	}
	if hasMetadata {
		endpoints[metadataEndpointPath] = "OAuth Protected Resource Metadata (RFC 9728)"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"name":      ServerName,
			"endpoints": endpoints,
		})
	}
}

// healthHandler is a lightweight liveness probe.
func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "healthy",
		})
	}
}

// limitBody caps request body size to prevent memory exhaustion from
// oversized payloads. Requests exceeding the limit receive a 413.
func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}
