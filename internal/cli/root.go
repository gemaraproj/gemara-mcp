// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"time"

	"github.com/gemaraproj/gemara-mcp/internal/tool"
	"github.com/gemaraproj/gemara-mcp/internal/tool/fetcher"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

const defaultCacheTTL = 1 * time.Hour

// New creates the root command
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gemara-mcp[command]",
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

func serveCmd() *cobra.Command {
	var modeName string

	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Start the Gemara MCP server",
		Example: "gemara-mcp serve\ngemara-mcp serve --mode advisory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cache := fetcher.NewCache(defaultCacheTTL)

			var (
				mode tool.Mode
				err  error
			)
			switch modeName {
			case "advisory":
				mode, err = tool.NewAdvisoryMode(cache)
			case "artifact":
				mode, err = tool.NewArtifactMode(cache)
			default:
				return fmt.Errorf("unknown mode %q: must be \"advisory\" or \"artifact\"", modeName)
			}
			if err != nil {
				return fmt.Errorf("initializing %s mode: %w", modeName, err)
			}

			server := mcp.NewServer(&mcp.Implementation{
				Name:    "gemara-mcp",
				Title:   "Gemara MCP",
				Version: GetVersion(),
			}, &mcp.ServerOptions{
				Instructions: mode.Description(),
			})

			mode.Register(server)

			return server.Run(cmd.Context(), &mcp.StdioTransport{})
		},
	}

	cmd.Flags().StringVar(&modeName, "mode", "artifact", "server mode: advisory (read-only) or artifact (advisory + guided artifact creation)")

	return cmd
}
