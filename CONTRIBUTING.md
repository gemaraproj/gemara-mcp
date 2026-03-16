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

### Building Docker Image

```bash
docker build --build-arg VERSION=$(git describe --tags --always) --build-arg BUILD=$(git rev-parse --short HEAD) -t gemara-mcp .
```

## Package Layout

| Package | Role |
|:---|:---|
| `cli` | Parses flags, creates the server mode, starts the MCP server |
| `server` | Defines MCP primitives (tools, resources, prompts) and operational modes |
| `server/fetcher` | Generic caching layer for remote data (HTTP, CUE registry) |
| `server/schema` | CUE schema loading, formatting, and validation |

## Reporting Issues

[Open an issue](https://github.com/gemaraproj/gemara-mcp/issues) on GitHub with a clear description of the bug or feature request.

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
