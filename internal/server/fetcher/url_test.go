// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{name: "valid semver", version: "v1.2.3", wantErr: false},
		{name: "valid semver with prerelease", version: "v1.0.0-rc.1", wantErr: false},
		{name: "valid latest", version: "latest", wantErr: false},
		{name: "missing v prefix", version: "1.2.3", wantErr: true},
		{name: "empty string", version: "", wantErr: true},
		{name: "path traversal", version: "../../../etc/passwd", wantErr: true},
		{name: "url injection", version: "v1.0.0?q=x", wantErr: true},
		{name: "fragment injection", version: "v1.0.0#frag", wantErr: true},
		{name: "slash injection", version: "v1.0.0/../../evil", wantErr: true},
		{name: "space injection", version: "v1.0.0 ", wantErr: true},
		{name: "arbitrary string", version: "https://evil.com", wantErr: true},
		{name: "null byte", version: "v1.0.0\x00", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewURLBuilder(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		suffix  string
		wantErr string
	}{
		{
			name:   "valid HTTPS base with path suffix",
			base:   "https://raw.githubusercontent.com/org/repo/",
			suffix: "/docs/file.yaml",
		},
		{
			name:   "valid HTTPS base without suffix",
			base:   "https://registry.example.com/pkg@",
			suffix: "",
		},
		{
			name:    "rejects HTTP",
			base:    "http://example.com/",
			suffix:  "",
			wantErr: "must use HTTPS",
		},
		{
			name:    "rejects empty scheme",
			base:    "example.com/path",
			suffix:  "",
			wantErr: "must use HTTPS",
		},
		{
			name:    "rejects missing host",
			base:    "https:///path",
			suffix:  "",
			wantErr: "must have a host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewURLBuilder(tt.base, tt.suffix)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, builder)
		})
	}
}

func TestURLBuilder_Build(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		suffix  string
		version string
		want    string
		wantErr bool
	}{
		{
			name:    "lexicon pattern with semver",
			base:    "https://raw.githubusercontent.com/gemaraproj/gemara/",
			suffix:  "/docs/lexicon.yaml",
			version: "v0.19.1",
			want:    "https://raw.githubusercontent.com/gemaraproj/gemara/v0.19.1/docs/lexicon.yaml",
		},
		{
			name:    "lexicon pattern with latest",
			base:    "https://raw.githubusercontent.com/gemaraproj/gemara/",
			suffix:  "/docs/lexicon.yaml",
			version: "latest",
			want:    "https://raw.githubusercontent.com/gemaraproj/gemara/latest/docs/lexicon.yaml",
		},
		{
			name:    "registry pattern with at-sign",
			base:    "https://registry.cue.works/docs/github.com/gemaraproj/gemara@",
			suffix:  "",
			version: "latest",
			want:    "https://registry.cue.works/docs/github.com/gemaraproj/gemara@latest",
		},
		{
			name:    "registry pattern with semver",
			base:    "https://registry.cue.works/docs/github.com/gemaraproj/gemara@",
			suffix:  "",
			version: "v1.0.0",
			want:    "https://registry.cue.works/docs/github.com/gemaraproj/gemara@v1.0.0",
		},
		{
			name:    "rejects invalid version",
			base:    "https://example.com/",
			suffix:  "",
			version: "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "rejects empty version",
			base:    "https://example.com/",
			suffix:  "",
			version: "",
			wantErr: true,
		},
		{
			name:    "rejects url as version",
			base:    "https://example.com/",
			suffix:  "",
			version: "https://evil.com/payload",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewURLBuilder(tt.base, tt.suffix)
			require.NoError(t, err)

			result, err := builder.Build(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}
