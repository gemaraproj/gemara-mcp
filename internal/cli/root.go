package cli

import (
	"fmt"

	"github.com/gemaraproj/gemara-mcp/internal/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// New creates the root command
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gemara-mcp[command]",
		SilenceUsage: true,
	}
	cmd.AddCommand(
		serveCmd,
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

var serveCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Start the Gemara MCP server",
	Example: "gemara-mcp serve",
	RunE: func(cmd *cobra.Command, args []string) error {
		advisory := tool.AdvisoryMode{}

		server := mcp.NewServer(&mcp.Implementation{
			Name:    "gemara-mcp",
			Title:   "Gemara MCP",
			Version: GetVersion(),
		}, &mcp.ServerOptions{
			Instructions: advisory.Description(),
		})

		advisory.Register(server)

		return server.Run(cmd.Context(), &mcp.StdioTransport{})
	},
}
