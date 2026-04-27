# gemara-mcp

Model Context Protocol server for Gemara artifact management.

## Quick Start

### stdio

MCP client spawns the server as a subprocess. Add to your MCP client configuration:

```json
{
  "mcpServers": {
    "gemara-mcp": {
      "command": "docker",
      "args": ["run", "--rm", "-i", "ghcr.io/gemaraproj/gemara-mcp:latest", "serve"]
    }
  }
}
```

Or with a local binary:

```json
{
  "mcpServers": {
    "gemara-mcp": {
      "command": "/absolute/path/to/gemara-mcp",
      "args": ["serve"]
    }
  }
}
```

### Streamable HTTP

Server runs independently; clients connect over the network.

```bash
gemara-mcp serve --transport http --insecure
```

```json
{
  "mcpServers": {
    "gemara-mcp": {
      "url": "http://127.0.0.1:8080/mcp"
    }
  }
}
```

## Configuration

| Flag                | Env Var                  | Default          | Description                                            |
|:--------------------|:-------------------------|:-----------------|:-------------------------------------------------------|
| `--mode`            | `GEMARA_MODE`            | `artifact`       | Server mode: `advisory` or `artifact`                  |
| `--transport`       | `GEMARA_TRANSPORT`       | `stdio`          | Transport: `stdio` or `http`                           |
| `--address`         | `GEMARA_ADDRESS`         | `127.0.0.1:8080` | HTTP listen address                                    |
| `--base-url`        | `GEMARA_BASE_URL`        |                  | Externally-reachable URL; overrides address in metadata |
| `--tls-cert`        | `GEMARA_TLS_CERT`        |                  | Path to TLS certificate file                           |
| `--tls-key`         | `GEMARA_TLS_KEY`         |                  | Path to TLS private key file                           |
| `--auth-server-url` | `GEMARA_AUTH_SERVER_URL` |                  | OAuth authorization server URL for RFC 9728 metadata   |
| `--required-scopes` | `GEMARA_REQUIRED_SCOPES` |                  | Comma-separated OAuth scopes advertised in metadata    |
| `--insecure`        | `GEMARA_INSECURE`        | `false`          | Allow HTTP transport without TLS or authentication     |

### Security Defaults

HTTP transport requires TLS and an `--auth-server-url`. The server will not start without them unless `--insecure` is passed.

The server does **not** validate tokens itself. Token validation must be handled by an upstream reverse proxy (Envoy, Istio, oauth2-proxy) or cloud load balancer. The `--auth-server-url` flag configures the RFC 9728 OAuth Protected Resource Metadata endpoint so MCP clients can discover where to obtain tokens.

Use `--insecure` for local development or when running behind a gateway that handles both TLS termination and token validation.

### Server Modes

| Mode       | Purpose                                                         |
|:-----------|:----------------------------------------------------------------|
| `advisory` | Read-only analysis and validation of existing artifacts         |
| `artifact` | All advisory capabilities plus guided artifact creation wizards |

## Container Deployment

### Building the Image

```bash
docker build --build-arg VERSION=$(git describe --tags --always) --build-arg BUILD=$(git rev-parse --short HEAD) -t gemara-mcp .
```

### Behind a Gateway

TLS termination and token validation handled by an upstream gateway or reverse proxy.

```bash
docker run -d --name gemara-mcp \
  -p 127.0.0.1:8080:8080 \
  ghcr.io/gemaraproj/gemara-mcp:latest \
  serve --transport http --address 0.0.0.0:8080 --base-url http://localhost:8080 --insecure
```

See `hack/docker-compose.yml` for a full gateway deployment example using Envoy for JWT validation.

## Capabilities

### Tools

| Tool                        | Mode     | Description                                            |
|:----------------------------|:---------|:-------------------------------------------------------|
| `validate_gemara_artifact`  | all      | Validate YAML content against Gemara CUE schema        |
| `migrate_gemara_artifact`   | artifact | Migrate a Gemara artifact from v0 to v1 schema         |

### Resources

| URI                                        | Description                                        |
|:-------------------------------------------|:---------------------------------------------------|
| `gemara://lexicon`                         | Term definitions for the Gemara security model     |
| `gemara://schema/definitions`              | CUE schema definitions for all artifact types      |
| `gemara://schema/definitions{?version}`    | CUE schema definitions for a specific version      |

### Prompts (artifact mode only)

| Prompt              | Description                                                    |
|:--------------------|:---------------------------------------------------------------|
| `threat_assessment` | Guided wizard for creating a Gemara-compatible Threat Catalog  |
| `control_catalog`   | Guided wizard for creating a Gemara-compatible Control Catalog |
| `migration`         | Guided wizard for migrating artifacts from v0 to v1 schema    |

## Verifying Image Signatures

Container images are signed with [cosign](https://docs.sigstore.dev/cosign/overview/) via GitHub Actions OIDC keyless signing.

```bash
cosign verify \
  --certificate-identity-regexp="https://github.com/gemaraproj/gemara-mcp/.github/workflows/release.yml" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  ghcr.io/gemaraproj/gemara-mcp@<DIGEST>
```
