// SPDX-License-Identifier: Apache-2.0

package httpserver

import (
	"fmt"
	"strings"
)

// Config holds HTTP-specific server configuration.
type Config struct {
	Address string
	BaseURL string
	TLSCert string
	TLSKey  string

	AuthServerURL  string
	RequiredScopes []string

	Insecure bool
}

// Validate checks that the security configuration is consistent.
// TLS and authentication are required unless Insecure is set.
func (c *Config) Validate() error {
	if c.Insecure {
		return nil
	}

	var errs []string

	tlsCert, tlsKey := c.TLSCert != "", c.TLSKey != ""
	switch {
	case tlsCert != tlsKey:
		errs = append(errs,
			"--tls-cert and --tls-key must be provided together")
	case !tlsCert && !tlsKey:
		errs = append(errs,
			"TLS is required for HTTP transport; provide --tls-cert and --tls-key, or pass --insecure to acknowledge the risk")
	}

	if c.AuthServerURL == "" {
		errs = append(errs,
			"authentication is required for HTTP transport; provide --auth-server-url, or pass --insecure to acknowledge the risk")
	}

	if len(errs) > 0 {
		return fmt.Errorf("security configuration error:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// ResourceURI builds the canonical resource identifier for RFC 9728 metadata.
// When BaseURL is set it takes precedence over the listen address, which
// is necessary for container deployments where the bind address (0.0.0.0)
// differs from the externally-reachable URL (localhost, a DNS name, etc.).
func ResourceURI(cfg *Config) string {
	return resolvedBase(cfg) + mcpEndpointPath
}

// MetadataURI builds the URL for the OAuth Protected Resource Metadata
// endpoint (RFC 9728). This is included in WWW-Authenticate headers on
// 401 responses so clients can discover where to obtain tokens.
func MetadataURI(cfg *Config) string {
	return resolvedBase(cfg) + metadataEndpointPath
}

func resolvedBase(cfg *Config) string {
	if cfg.BaseURL != "" {
		return strings.TrimRight(cfg.BaseURL, "/")
	}
	return fmt.Sprintf("%s://%s", resourceScheme(cfg), cfg.Address)
}

func resourceScheme(cfg *Config) string {
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		return "https"
	}
	return "http"
}
