# Contributing to gemara-mcp

The project welcomes your contributions whether they be:

* Reporting an [issue](https://github.com/gemaraproj/gemara-mcp/issues/new/choose)
* Making a code contribution ([create a fork](https://github.com/gemaraproj/gemara-mcp/fork))
* Updating our [docs](https://github.com/gemaraproj/gemara-mcp/blob/main/README.md)

## PR guidelines

All changes to the repository should be made via PR ([OSPS-AC-03](https://baseline.openssf.org/#osps-ac-03)).

PRs MUST meet the following criteria:

* Clear title that conforms to the [Conventional Commits spec](https://www.conventionalcommits.org/)
* Descriptive commit message
* DCO signoff (via `git commit -s` -- [OSPS-LE-01](https://baseline.openssf.org/#osps-le-01))
* All checks must pass ([OSPS-QA-04](https://baseline.openssf.org/#osps-qa-04))

## Development Workflow

1. Fork and clone the repository.
2. Create a feature branch from `main`.
3. Run `make build` to verify the project compiles.

| Command      | Purpose                |
|:-------------|:-----------------------|
| `make build` | Build the binary       |
| `make test`  | Run tests              |
| `make vet`   | Run `go vet`           |
| `make fmt`   | Format code            |
| `make lint`  | Run all linting checks |

Run `make ci` to execute the full check suite before submitting a PR.

## Testing

### Unit Tests

```bash
make test
```

Run a specific package:

```bash
go test -v ./internal/httpserver/...
```

### Integration Tests

Integration tests are gated behind the `integration` build tag and require network access (e.g. fetching CUE schemas from the OCI registry).

```bash
make test-integration
```

### Coverage

```bash
make test-coverage
```

Generates `coverage.html` in the project root.

### Docker Compose Auth Stacks

The `hack/` directory contains docker-compose profiles for testing the Streamable HTTP transport with authentication. The server delegates token validation to an upstream gateway proxy.

| Command | Description |
|:---|:---|
| `make compose-gateway` | Mock IdP — Envoy validates JWTs at the edge |
| `make compose-hydra` | Ory Hydra — DCR-enabled IdP for real MCP client OAuth testing |
| `make compose-token` | Fetch a test token from the mock IdP |
| `make compose-down` | Tear down all compose services and volumes |

Gateway smoke test:

```bash
make compose-gateway
TOKEN=$(make compose-token)
curl -X POST http://localhost:7070/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}'
```

### Testing with an MCP Client

Build the container image and point your MCP client at it to test end-to-end.

```bash
docker build --build-arg VERSION=$(git describe --tags --always) --build-arg BUILD=$(git rev-parse --short HEAD) -t gemara-mcp .
```

#### stdio

Add the server to your MCP client configuration. Most clients accept the stdio transport via a JSON config block:

```json
{
  "mcpServers": {
    "gemara-mcp": {
      "command": "docker",
      "args": ["run", "--rm", "-i", "gemara-mcp", "serve"]
    }
  }
}
```

Restart the client (or reload the MCP server list) after updating the config. The client spawns the container automatically.

#### Streamable HTTP

Start the container, then configure the client to connect over HTTP:

```bash
docker run --rm -p 8080:8080 gemara-mcp \
  serve --transport http --insecure --address 0.0.0.0:8080 --base-url http://localhost:8080
```

```json
{
  "mcpServers": {
    "gemara-mcp": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

#### OAuth with an MCP client

MCP clients that support OAuth 2.1 with [RFC 9728](https://datatracker.ietf.org/doc/html/rfc9728) discovery can authenticate against a remote server automatically. The OAuth flow requires the authorization server to support Dynamic Client Registration (DCR).

The `hydra` compose profile runs [Ory Hydra](https://github.com/ory/hydra) (Apache-2.0) with DCR enabled and JWT access tokens. Start it, then point your MCP client at the endpoint:

```bash
make compose-hydra
```

```json
{
  "mcpServers": {
    "gemara-mcp": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

The client discovers the authorization server via `/.well-known/oauth-protected-resource`, registers itself via DCR, and redirects you to a browser login page (any username works). After login, the client exchanges the authorization code for a JWT and connects to the MCP endpoint.

Tear down after testing:

```bash
make compose-down
```

The `gateway` profile uses `navikt/mock-oauth2-server`, which does not support DCR. Use that profile for curl-based auth testing (see [Docker Compose Auth Stacks](#docker-compose-auth-stacks)).

#### Verifying connectivity

Once connected, confirm the client can reach the server by:

1. Listing available tools — `validate_gemara_artifact` and `migrate_gemara_artifact` should appear.
2. Reading `gemara://lexicon` — should return the Gemara term definitions.
3. Running a prompt — `threat_assessment` or `control_catalog` (artifact mode only).

### Building Docker Image

```bash
docker build --build-arg VERSION=$(git describe --tags --always) --build-arg BUILD=$(git rev-parse --short HEAD) -t gemara-mcp .
```

## Package Layout

| Package | Role |
|:---|:---|
| `cli` | Parses flags, creates the server mode, starts the MCP server |
| `httpserver` | Streamable HTTP transport with RFC 9728 metadata and gateway auth |
| `server` | Defines MCP primitives (tools, resources, prompts) and operational modes |
| `server/fetcher` | Generic caching layer for remote data (HTTP, CUE registry) |
| `server/schema` | CUE schema loading, formatting, and validation |

## Reporting Issues

[Open an issue](https://github.com/gemaraproj/gemara-mcp/issues) on GitHub with a clear description of the bug or feature request.

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
