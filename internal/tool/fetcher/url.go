// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"fmt"
	"net/url"
	"regexp"
)

// URLBuilder constructs fetch URLs from base components and a
// version string.
type URLBuilder struct {
	base   *url.URL
	suffix string
}

// NewURLBuilder creates a URLBuilder from base URL and path suffix.
func NewURLBuilder(baseURL, suffix string) (*URLBuilder, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("base URL must use HTTPS, got scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("base URL must have a host")
	}
	return &URLBuilder{base: u, suffix: suffix}, nil
}

// Build constructs a full URL by inserting the validated version between
// the trusted and suffix.
func (b *URLBuilder) Build(version string) (string, error) {
	if err := ValidateVersion(version); err != nil {
		return "", err
	}

	raw := b.base.String() + version + b.suffix

	result, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("constructed invalid URL: %w", err)
	}

	if result.Scheme != b.base.Scheme || result.Host != b.base.Host {
		return "", fmt.Errorf(
			"constructed URL has unexpected origin: got %s://%s, want %s://%s",
			result.Scheme, result.Host, b.base.Scheme, b.base.Host,
		)
	}

	return result.String(), nil
}

var validVersionPattern = regexp.MustCompile(`^(v\d+\.\d+\.\d+(-[\w.]+)?|latest)$`)

// ValidateVersion checks that a version string matches semver (e.g., v1.2.3) or "latest".
func ValidateVersion(version string) error {
	if !validVersionPattern.MatchString(version) {
		return fmt.Errorf(
			"invalid version %q: must be semver (e.g., v1.2.3) or %q",
			version, "latest",
		)
	}
	return nil
}
