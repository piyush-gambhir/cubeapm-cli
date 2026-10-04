# Contributing to CubeAPM CLI

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.26 or later (Go 1.27.1 recommended)
- Make
- Git

### Clone and Build

```bash
git clone https://github.com/piyush-gambhir/cubeapm-cli.git
cd cubeapm-cli/cli-go
make build
```

### Run Locally

```bash
./cubeapm --help
./cubeapm version
```

### Run Tests

```bash
make test
```

### Lint

```bash
make lint    # requires golangci-lint
make vet     # go vet
make fmt     # gofmt
```

## Project Structure

```
.
├── cli-go/                     # Go module for the `cubeapm` binary
│   ├── main.go                 # Entry point
│   ├── VERSION                 # Release version (bump before tagging)
│   ├── Makefile                # build, install, test, lint, fmt, vet, llms-check
│   ├── .goreleaser.yaml        # Release archives (tar.gz, zip on Windows), checksums, SBOMs
│   ├── cmd/                    # Cobra command definitions
│   │   ├── root.go             # Root command, global flags, read-only and no-input checks
│   │   ├── login.go            # Interactive login
│   │   ├── version.go          # version command
│   │   ├── update.go           # update command (self-update on macOS, Linux, Windows; --check)
│   │   ├── traces/             # search, get, services, operations, dependencies, callers
│   │   ├── metrics/            # query, query-range, labels, label-values, series
│   │   ├── logs/               # query, hits, stats, streams, field-names, field-values, status
│   │   │   └── delete/         # Log deletion tasks: list, run, stop
│   │   ├── ingest/             # Data ingestion: metrics, logs
│   │   └── config/             # view, get, set, profiles (list, use, delete)
│   ├── internal/
│   │   ├── auth/               # Ory Kratos login flow (session cookie)
│   │   ├── client/             # HTTP API client
│   │   │   ├── client.go       # Base client (auth, headers, multi-port routing)
│   │   │   ├── traces.go       # Jaeger-compatible trace API methods
│   │   │   ├── metrics.go      # Prometheus-compatible metrics API methods
│   │   │   ├── logs.go         # VictoriaLogs-compatible log API methods
│   │   │   ├── ingest.go       # Ingest API methods (metrics push, log push)
│   │   │   ├── admin.go        # Admin API methods (log deletion tasks)
│   │   │   └── response_limit.go # Response size limits
│   │   ├── cmdutil/            # Shared command state (config, client, output format, flags)
│   │   ├── config/             # Config file, profiles, and auth resolution
│   │   ├── output/             # JSON/YAML/Table formatters
│   │   ├── timeflag/           # Flexible time range parsing (RFC3339, Unix, relative durations)
│   │   ├── types/              # Shared data types (trace, metric, log, common)
│   │   └── update/             # Release check, update notice, and self-update
│   └── scripts/
│       ├── check-llms.sh       # Checks docs/llms.txt documents every command
│       └── deploy.sh           # Local GoReleaser wrapper
├── cli-mcp-server/             # MCP server that exposes the CLI to MCP clients
├── cubeapm/                    # Claude Code skill (SKILL.md, references/commands.md)
├── docs/                       # llms.txt, COMPATIBILITY.md, CREDENTIALS.md
├── web/                        # Docs site (Next.js and Fumadocs on Cloudflare Workers)
├── scripts/deploy-docs.sh      # Builds and deploys the docs site
├── install.sh                  # Install script for macOS and Linux
└── .github/workflows/
    ├── ci.yml                  # Go, docs site, and MCP server checks on main pushes and PRs
    ├── codeql.yml              # CodeQL analysis
    └── release.yml             # GoReleaser on tag push
```

### Multi-Port Architecture

CubeAPM uses three ports for different purposes. The client in `internal/client/client.go` routes requests to the correct port based on the operation:

| Port | Default | Purpose |
|------|---------|---------|
| Query port | 3140 | Read queries: traces, metrics, logs |
| Ingest port | 3130 | Write/push: metrics ingestion, log ingestion |
| Admin port | 3199 | Administration: log deletion tasks |

When adding new API methods, make sure to use the correct port for the operation type.

### Time Flag Parsing

The `internal/timeflag/` package handles flexible time range parsing across all query commands. It supports:

- Relative durations (`1h`, `30m`, `2d`, `1d12h`)
- RFC3339 timestamps (`2024-01-15T10:00:00Z`)
- Unix timestamps (`1705312800`)
- Relative from/to (`-2h`, `-1h`)

Use the shared `timeflag` helpers when adding new commands that accept time ranges.

### Data Types

The `internal/types/` package defines shared data structures for traces (Jaeger format), metrics (Prometheus vector/matrix), and logs (VictoriaLogs format). Use these types when adding new API methods or formatters.

## Adding a New Command

Paths are relative to `cli-go/`.

1. **Add the API method** in `internal/client/<resource>.go`:
   ```go
   func (c *Client) ListWidgets(params ...) ([]Widget, error) {
       // HTTP call to the CubeAPM API
   }
   ```

2. **Create the command** in `cmd/<resource>/list.go`:
   ```go
   func newListCmd() *cobra.Command {
       // Define flags, run function, help text with examples.
       // Use cmdutil.APIClient and cmdutil.OutputFormat, set by the root command.
   }
   ```

3. **Register** the command in the parent command's `New*Cmd()` function.

4. **Add a test** in the corresponding `_test.go` file using `httptest.NewServer`.

5. **Update documentation**:
   - Add a `Long` description with examples to the command
   - Update `README.md` with the new command
   - Update `CLAUDE.md` if it's a commonly-used command
   - Update the skill's `references/commands.md`

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable names
- Every command must have:
  - `Short` description (one line)
  - `Long` description with usage examples
  - Proper flag definitions with descriptions
- Use `-o json` output in all examples for agent-friendliness
- Table output should have meaningful column headers

## Commit Messages

Follow conventional commits:
```
feat: add widget list command
fix: correct pagination in dashboard search
docs: update README with new alert commands
test: add tests for credential CRUD
chore: update dependencies
```

## Pull Requests

1. Fork the repo and create a feature branch
2. Make your changes with tests
3. Run `make test` and `make vet` to ensure everything passes
4. Commit with a clear message
5. Open a PR against `main`

## Releasing

Releases are automated via GoReleaser. To create a release, bump `cli-go/VERSION`
in a pull request, then tag the merge commit on `main` with signed tags. The
`cli-go/` tag versions the Go module, which lives in the `cli-go/` subdirectory.

```bash
git fetch origin
commit=MERGE_COMMIT_SHA  # the merge commit of the version bump PR
v="$(git show "$commit:cli-go/VERSION")"
git tag -s "v$v" -m "v$v" "$commit"
git tag -s "cli-go/v$v" -m "cli-go/v$v" "$commit"
git push origin "v$v" "cli-go/v$v"
```

The `v*` tag triggers GitHub Actions to:
1. Build binaries for all platforms
2. Create a GitHub Release with assets
3. Generate a changelog

## Reporting Issues

- Use GitHub Issues
- Include: CLI version (`cubeapm version`), OS/arch, command that failed, error output
- For feature requests, describe the use case

## License

This project is licensed under the MIT License, see [LICENSE](LICENSE) for details.
