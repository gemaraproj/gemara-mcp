// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"time"

	"github.com/gemaraproj/gemara-mcp/internal/server"
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

	cmd.Flags().StringVar(&modeName, "mode", "artifact", "server mode: advisory (consumer, read-only evaluation) or artifact (producer, guided artifact creation)")

	return cmd
}
