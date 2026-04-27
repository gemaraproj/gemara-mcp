// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gemaraproj/gemara-mcp/internal/httpserver"
	"github.com/gemaraproj/gemara-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"os"
	"strings"
)

const defaultCacheTTL = 1 * time.Hour

// serveConfig holds all flags for the serve command.
type serveConfig struct {
	Mode      string
	Transport string
	HTTP      httpserver.Config
}

// New creates the root command
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gemara-mcp [command]",
		SilenceUsage: true,
	}
	cmd.AddCommand(
		serveCmd(),
		versionCmd,
	)
	return cmd
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Gemara MCP Server %s\n", GetVersion())
	},
}

// bindEnvDefaults walks every flag on cmd and, for each that was not
// explicitly set on the command line, checks for an environment variable
// named PREFIX_FLAG_NAME (kebab → UPPER_SNAKE).
func bindEnvDefaults(cmd *cobra.Command, prefix string) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		envKey := prefix + "_" + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if val, ok := os.LookupEnv(envKey); ok && !f.Changed {
			_ = f.Value.Set(val)
		}
	})
}

func serveCmd() *cobra.Command {
	var cfg serveConfig

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Gemara MCP server",
		Example: `  gemara-mcp serve
  gemara-mcp serve --mode advisory
  gemara-mcp serve --transport http --insecure
  gemara-mcp serve --transport http --tls-cert cert.pem --tls-key key.pem --auth-server-url https://auth.example.com/`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			bindEnvDefaults(cmd, "GEMARA")
			switch cfg.Transport {
			case "stdio":
			case "http":
				if err := cfg.HTTP.Validate(); err != nil {
					return fmt.Errorf("http config error: %w", err)
				}
				if cfg.HTTP.Insecure {
					slog.Warn("--insecure: TLS and authentication disabled", "address", cfg.HTTP.Address)
				} else {
					slog.Warn("gateway mode: /mcp is NOT authenticated at the application layer — an upstream proxy (Envoy, oauth2-proxy) MUST validate tokens before traffic reaches this server",
						"auth_server", cfg.HTTP.AuthServerURL)
					if len(cfg.HTTP.RequiredScopes) > 0 {
						slog.Warn("--required-scopes are advertised in metadata but only enforced by the upstream proxy")
					}
				}
			default:
				return fmt.Errorf("unknown transport %q: must be \"stdio\" or \"http\"", cfg.Transport)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			mcpServer, err := newMCPServer(cfg.Mode)
			if err != nil {
				return err
			}
			switch cfg.Transport {
			case "http":
				return httpserver.Run(cmd.Context(), mcpServer, &cfg.HTTP)
			default:
				return mcpServer.Run(cmd.Context(), &mcp.StdioTransport{})
			}
		},
	}

	cmd.Flags().StringVar(&cfg.Mode, "mode", "artifact",
		"server mode: advisory or artifact")
	cmd.Flags().StringVar(&cfg.Transport, "transport", "stdio",
		"transport: stdio or http (env: GEMARA_TRANSPORT)")
	cmd.Flags().StringVar(&cfg.HTTP.Address, "address", "127.0.0.1:8080",
		"HTTP listen address; use 0.0.0.0 inside containers (env: GEMARA_ADDRESS)")
	cmd.Flags().StringVar(&cfg.HTTP.BaseURL, "base-url", "",
		"externally-reachable base URL; overrides address in RFC 9728 metadata (env: GEMARA_BASE_URL)")
	cmd.Flags().StringVar(&cfg.HTTP.TLSCert, "tls-cert", "",
		"path to TLS certificate file (env: GEMARA_TLS_CERT)")
	cmd.Flags().StringVar(&cfg.HTTP.TLSKey, "tls-key", "",
		"path to TLS private key file (env: GEMARA_TLS_KEY)")

	cmd.Flags().StringVar(&cfg.HTTP.AuthServerURL, "auth-server-url", "",
		"OAuth authorization server URL for RFC 9728 metadata (env: GEMARA_AUTH_SERVER_URL)")
	cmd.Flags().StringSliceVar(&cfg.HTTP.RequiredScopes, "required-scopes", nil,
		"comma-separated OAuth scopes advertised in metadata (env: GEMARA_REQUIRED_SCOPES)")

	cmd.Flags().BoolVar(&cfg.HTTP.Insecure, "insecure", false,
		"allow HTTP transport without TLS and authentication (env: GEMARA_INSECURE)")

	return cmd
}

// newMCPServer creates and configures the MCP server for the given mode.
func newMCPServer(modeName string) (*mcp.Server, error) {
	var (
		mode server.Mode
		err  error
	)
	switch modeName {
	case "advisory":
		mode, err = server.NewAdvisoryMode(defaultCacheTTL)
	case "artifact":
		mode, err = server.NewArtifactMode(defaultCacheTTL)
	default:
		return nil, fmt.Errorf("unknown mode %q: must be \"advisory\" or \"artifact\"", modeName)
	}
	if err != nil {
		return nil, fmt.Errorf("initializing %s mode: %w", modeName, err)
	}

	s := mcp.NewServer(&mcp.Implementation{
		Name:    httpserver.ServerName,
		Title:   "Gemara MCP",
		Version: GetVersion(),
	}, &mcp.ServerOptions{
		Instructions: mode.Description(),
	})

	mode.Register(s)
	return s, nil
}
