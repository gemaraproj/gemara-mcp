# gemara-mcp

Gemara MCP Server - A Model Context Protocol server for Gemara artifact management.

## Building

Build the binary:

```bash
make build
```

## Installation

### MCP Client Configuration

To use this server with an MCP client, add it to your MCP configuration file.

Add the following configuration (adjust the path to your binary):

```json
{
  "mcpServers": {
    "gemara-mcp": {
      "command": "/absolute/path/to/gemara-mcp/bin/gemara-mcp",
      "args": ["serve"]
    }
  }
}
```

#### Using Docker

If running from Docker, use:

```json
{
  "mcpServers": {
    "gemara-mcp": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "gemara-mcp:latest",
        "serve"
      ]
    }
  }
}
```

## Available Tools

The server provides read-only information about Gemara artifacts in the workspace.

- **get_lexicon**: Retrieve Gemara lexicon entries
- **validate_gemara_artifact**: Validate YAML artifacts against Gemara schema definitions

## Available Resources

- **gemara://lexicon**: Access the Gemara lexicon as a resource

### Building Docker Image

```bash
docker build --build-arg VERSION=$(git describe --tags --always) --build-arg BUILD=$(git rev-parse --short HEAD) -t gemara-mcp .
```
