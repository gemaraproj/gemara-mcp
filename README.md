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
        "ghcr.io/gemaraproj/gemara-mcp:latest",
        "serve"
      ]
    }
  }
}
```

## Server Modes

The server operates in one of two modes, selected with the `--mode` flag (default: `artifact`).

| Mode | Purpose |
|:---|:---|
| `advisory` | Read-only analysis and validation of existing artifacts |
| `artifact` | All advisory capabilities plus guided artifact creation wizards |

```bash
gemara-mcp serve --mode advisory
gemara-mcp serve --mode artifact
```

## Available Tools, Resources, and Prompts

### Tools

| Tool | Description |
|:---|:---|
| `validate_gemara_artifact` | Validate YAML content against Gemara CUE schema definitions |

### Resources

| Resource URI | Description |
|:---|:---|
| `gemara://lexicon` | Term definitions for the Gemara security model |
| `gemara://schema/definitions` | CUE schema definitions for all Gemara artifact types (latest version) |
| `gemara://schema/definitions{?version}` | CUE schema definitions for a specific Gemara module version |

### Prompts (artifact mode only)

| Prompt | Description |
|:---|:---|
| `threat_assessment` | Interactive wizard for creating a Gemara-compatible Threat Catalog |
| `control_catalog` | Interactive wizard for creating a Gemara-compatible Control Catalog |

## Verifying Image Signatures

Released container images are signed with [cosign](https://docs.sigstore.dev/cosign/overview/) using keyless signing via GitHub Actions OIDC. Signatures are attached to the image digest, so verification works with any tag.

```bash
cosign verify \
  --certificate-identity-regexp="https://github.com/gemaraproj/gemara-mcp/.github/workflows/release.yml" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  ghcr.io/gemaraproj/gemara-mcp:latest
```

## Building Docker Image

```bash
docker build --build-arg VERSION=$(git describe --tags --always) --build-arg BUILD=$(git rev-parse --short HEAD) -t gemara-mcp .
```
