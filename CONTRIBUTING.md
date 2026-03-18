# Contributing to CubeAPM CLI

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.22 or later
- Make
- Git

### Clone and Build

```bash
git clone https://github.com/piyush-gambhir/cubeapm-cli.git
cd cubeapm-cli
make build
```

### Run Locally

```bash
./bin/cubeapm --help
./bin/cubeapm version
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
├── main.go                 # Entry point
├── cmd/                    # Cobra command definitions
│   ├── root.go             # Root command, global flags
│   ├── login.go            # Auth commands
│   ├── traces/             # Trace query commands (search, get, services, operations, dependencies)
│   ├── metrics/            # Metrics commands (query, query-range, labels, label-values, series)
│   ├── logs/               # Log commands (query, hits, stats, streams, field-names, field-values, delete)
│   ├── ingest/             # Data ingestion commands (metrics, logs)
│   └── config/             # CLI config management (view, set, get, profiles)
├── internal/
│   ├── client/             # HTTP API client
│   │   ├── client.go       # Base client (auth, headers, multi-port routing)
│   │   ├── traces.go       # Jaeger-compatible trace API methods
│   │   ├── metrics.go      # Prometheus-compatible metrics API methods
│   │   ├── logs.go         # VictoriaLogs-compatible log API methods
│   │   ├── ingest.go       # Ingest API methods (metrics push, log push)
│   │   ├── admin.go        # Admin API methods (log deletion tasks)
│   │   └── ...
│   ├── cmdutil/            # Shared command utilities (Factory, flag helpers)
│   ├── config/             # Config file and auth resolution
│   ├── output/             # JSON/YAML/Table formatters
│   ├── timeflag/           # Flexible time range parsing (RFC3339, Unix, relative durations)
│   ├── types/              # Shared data types
│   │   ├── trace.go        # Trace and span types
│   │   ├── metric.go       # Metric result types (vector, matrix)
│   │   ├── log.go          # Log entry and stream types
│   │   └── common.go       # Common types
│   └── update/             # Self-update logic
├── Makefile
├── .goreleaser.yaml
└── .github/workflows/
    ├── ci.yml              # Build + test on every push/PR
    └── release.yml         # GoReleaser on tag push
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

1. **Add the API method** in `internal/client/<resource>.go`:
   ```go
   func (c *Client) ListWidgets(params ...) ([]Widget, error) {
       // HTTP call to the CubeAPM API
   }
   ```

2. **Create the command** in `cmd/<resource>/list.go`:
   ```go
   func NewListCmd(f *cmdutil.Factory) *cobra.Command {
       // Define flags, run function, help text with examples
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

Releases are automated via GoReleaser. To create a release:

```bash
git tag v0.2.0
git push origin v0.2.0
```

This triggers GitHub Actions to:
1. Build binaries for all platforms
2. Create a GitHub Release with assets
3. Generate a changelog

## Reporting Issues

- Use GitHub Issues
- Include: CLI version (`cubeapm version`), OS/arch, command that failed, error output
- For feature requests, describe the use case

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.
