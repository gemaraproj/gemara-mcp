package cli

// Version information
// These can be set via ldflags during build:
// -X github.com/gemaraproj/gemara-mcp/internal/cli.Version=...
// -X github.com/gemaraproj/gemara-mcp/internal/cli.Build=...
var (
	Version = "0.1.0"
	Build   = "dev"
)

// GetVersion returns the version string
func GetVersion() string {
	return Version + "-" + Build
}
