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
| `migrate_gemara_artifact` | Migrate a Gemara artifact to v1 schema using CUE transformations |

### Resources

| Resource URI | Description |
|:---|:---|
| `gemara://about` | LLM-optimized overview of Gemara: the seven-layer model, artifact types, schemas, and SDKs. Fetched from the website with an embedded offline fallback |
| `gemara://lexicon` | Term definitions for the Gemara security model |
| `gemara://schema/definitions` | CUE schema definitions for all Gemara artifact types (latest version) |
| `gemara://schema/definitions{?version}` | CUE schema definitions for a specific Gemara module version |

### Prompts (artifact mode only)

| Prompt | Description |
|:---|:---|
| `threat_assessment` | Interactive wizard for creating a Gemara-compatible Threat Catalog |
| `control_catalog` | Interactive wizard for creating a Gemara-compatible Control Catalog |
| `migration` | Interactive wizard that guides you through migrating Gemara artifacts from v0 to v1 schema |


## Verifying Artifacts

Released artifacts include [SLSA build provenance](https://slsa.dev/provenance/v1) attestations generated via GitHub Actions OIDC. Binary archives also include SPDX SBOMs, each with its own attestation.

### Container images

```bash
gh attestation verify oci://ghcr.io/gemaraproj/gemara-mcp:<TAG> \
  --repo gemaraproj/gemara-mcp
```

### Binary archives

```bash
gh attestation verify gemara-mcp_Linux_x86_64.tar.gz \
  --repo gemaraproj/gemara-mcp
```

## Building Docker Image

```bash
docker build --build-arg VERSION=$(git describe --tags --always) --build-arg BUILD=$(git rev-parse --short HEAD) -t gemara-mcp .
```
