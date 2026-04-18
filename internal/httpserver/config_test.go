// SPDX-License-Identifier: Apache-2.0

package httpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "no TLS no auth errors",
			cfg:     Config{},
			wantErr: "security configuration error",
		},
		{
			name: "no TLS errors even with auth",
			cfg: Config{
				AuthServerURL: "https://auth.example.com/",
			},
			wantErr: "TLS is required",
		},
		{
			name: "no auth errors even with TLS",
			cfg: Config{
				TLSCert: "cert.pem",
				TLSKey:  "key.pem",
			},
			wantErr: "authentication is required",
		},
		{
			name: "TLS and auth passes",
			cfg: Config{
				TLSCert:       "cert.pem",
				TLSKey:        "key.pem",
				AuthServerURL: "https://auth.example.com/",
			},
		},
		{
			name: "insecure bypasses all requirements",
			cfg: Config{
				Insecure: true,
			},
		},
		{
			name: "partial TLS cert only errors",
			cfg: Config{
				TLSCert:       "cert.pem",
				AuthServerURL: "https://auth.example.com/",
			},
			wantErr: "must be provided together",
		},
		{
			name: "partial TLS key only errors",
			cfg: Config{
				TLSKey:        "key.pem",
				AuthServerURL: "https://auth.example.com/",
			},
			wantErr: "must be provided together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestResourceURI(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "HTTP without TLS",
			cfg:  Config{Address: "127.0.0.1:8080"},
			want: "http://127.0.0.1:8080/mcp",
		},
		{
			name: "HTTPS with TLS",
			cfg: Config{
				Address: "example.com:443",
				TLSCert: "cert.pem",
				TLSKey:  "key.pem",
			},
			want: "https://example.com:443/mcp",
		},
		{
			name: "partial TLS config treated as HTTP",
			cfg: Config{
				Address: "127.0.0.1:8080",
				TLSCert: "cert.pem",
			},
			want: "http://127.0.0.1:8080/mcp",
		},
		{
			name: "BaseURL overrides address",
			cfg: Config{
				Address: "0.0.0.0:8080",
				BaseURL: "http://localhost:8080",
			},
			want: "http://localhost:8080/mcp",
		},
		{
			name: "BaseURL trailing slash stripped",
			cfg: Config{
				Address: "0.0.0.0:8080",
				BaseURL: "https://api.example.com/",
			},
			want: "https://api.example.com/mcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ResourceURI(&tt.cfg))
		})
	}
}

func TestMetadataURI(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "HTTP without TLS",
			cfg:  Config{Address: "127.0.0.1:8080"},
			want: "http://127.0.0.1:8080/.well-known/oauth-protected-resource",
		},
		{
			name: "HTTPS with TLS",
			cfg: Config{
				Address: "example.com:443",
				TLSCert: "cert.pem",
				TLSKey:  "key.pem",
			},
			want: "https://example.com:443/.well-known/oauth-protected-resource",
		},
		{
			name: "BaseURL overrides address",
			cfg: Config{
				Address: "0.0.0.0:8080",
				BaseURL: "http://localhost:8080",
			},
			want: "http://localhost:8080/.well-known/oauth-protected-resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MetadataURI(&tt.cfg))
		})
	}
}
