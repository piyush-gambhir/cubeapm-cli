# CubeAPM CLI - Agent Guide

## Quick Reference

- **Binary:** `cubeapm`
- **Config file:** `~/.config/cubeapm/config.yaml`
- **Env vars:** `CUBEAPM_SERVER`, `CUBEAPM_TOKEN`, `CUBEAPM_QUERY_PORT`, `CUBEAPM_INGEST_PORT`, `CUBEAPM_ADMIN_PORT`
- **Config priority:** CLI flags > environment variables > profile config
- **Query language:** PromQL (metrics), LogsQL (logs), Jaeger format (traces)

## Multi-Port Architecture

CubeAPM uses three ports for different purposes:

| Port | Default | Purpose |
|------|---------|---------|
| Query port | 3140 | Read queries: traces, metrics, logs |
| Ingest port | 3130 | Write/push: metrics ingestion, log ingestion |
| Admin port | 3199 | Administration: log deletion tasks |

Override via flags (`--query-port`, `--ingest-port`, `--admin-port`), env vars, or profile config.

## Setup

```bash
# Interactive login (prompts for server, token, ports, profile name)
cubeapm login

# Or set environment variables for non-interactive use
export CUBEAPM_SERVER=cubeapm.example.com
export CUBEAPM_TOKEN=your-api-token
```

## Output Formats

All query commands support three output formats via `-o`:

- `-o table` (default) -- human-readable tabular output
- `-o json` -- JSON, ideal for programmatic parsing with jq
- `-o yaml` -- YAML, useful for config management

**For agents:** Always use `-o json` when you need to parse or process output programmatically.

## Time Range Notation

Most query commands accept time range flags. There are three ways to specify them:

| Method | Flags | Examples |
|--------|-------|---------|
| Relative | `--last <duration>` | `--last 1h`, `--last 30m`, `--last 2d`, `--last 1d12h` |
| Absolute (RFC3339) | `--from`, `--to` | `--from 2024-01-15T10:00:00Z --to 2024-01-15T12:00:00Z` |
| Absolute (Unix) | `--from`, `--to` | `--from 1705312800 --to 1705356000` |

If no time flags are provided, the default is the last 1 hour.

## Common Workflows

### Investigate errors end-to-end (traces, metrics, logs)

```bash
# 1. Find which services are reporting traces
cubeapm traces services -o json

# 2. Search for error traces in a service
cubeapm traces search --service api-gateway --status error --last 1h -o json

# 3. Get the full trace details with waterfall view
cubeapm traces get <trace-id>

# 4. Check error rate metrics for the service
cubeapm metrics query 'rate(http_requests_total{service="api-gateway",status=~"5.."}[5m])'

# 5. Search logs around the same time
cubeapm logs query 'error' --service api-gateway --last 1h -o json
```

### Find slow endpoints

```bash
# Search for traces slower than 500ms
cubeapm traces search --service api-gateway --min-duration 500ms --last 1h -o json

# Check p99 latency via metrics
cubeapm metrics query 'histogram_quantile(0.99, sum by (le) (rate(http_duration_seconds_bucket{service="api-gateway"}[5m])))'

# Check latency trend over time
cubeapm metrics query-range 'histogram_quantile(0.99, sum by (le) (rate(http_duration_seconds_bucket{service="api-gateway"}[5m])))' --last 6h --step 5m -o json
```

### Explore service dependencies

```bash
# Show service dependency graph for the last 24 hours
cubeapm traces dependencies --last 24h -o json

# Export as Graphviz DOT format and render to PNG
cubeapm traces dependencies --last 24h --dot | dot -Tpng -o deps.png
```

### Search and analyze traces

```bash
# Search traces by service in the last hour
cubeapm traces search --service api-gateway --last 1h -o json

# Search with multiple filters
cubeapm traces search --service payments --status error --min-duration 500ms --last 2h -o json

# Filter by operation name
cubeapm traces search --service api-gateway --query "GET /api/users" --last 1h -o json

# Filter by span tags
cubeapm traces search --service api-gateway --tags "http.method=POST" --tags "http.status_code=500" -o json

# Filter by environment and span kind
cubeapm traces search --service payments --env production --span-kind server -o json

# Search with a custom time range
cubeapm traces search --service auth --from 2024-01-15T00:00:00Z --to 2024-01-15T12:00:00Z -o json

# Return more results (default limit is 20)
cubeapm traces search --service api-gateway --limit 100 -o json

# Get a specific trace by ID (table mode shows waterfall view)
cubeapm traces get abc123def456789

# Get trace as raw JSON for scripting
cubeapm traces get abc123def456789 -o json

# List all operations for a service
cubeapm traces operations api-gateway -o json

# List only server-side operations
cubeapm traces operations api-gateway --span-kind server -o json
```

### Query metrics with PromQL

```bash
# Simple instant query (returns current value)
cubeapm metrics query 'up' -o json

# Request rate per service
cubeapm metrics query 'sum by (service) (rate(http_requests_total[5m]))' -o json

# Error rate as a percentage
cubeapm metrics query 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100'

# P99 latency
cubeapm metrics query 'histogram_quantile(0.99, sum by (le) (rate(http_duration_seconds_bucket[5m])))'

# Query at a specific time
cubeapm metrics query 'up' --time 2024-01-15T10:00:00Z -o json

# Range query: request rate over the last hour
cubeapm metrics query-range 'rate(http_requests_total[5m])' --last 1h --step 1m -o json

# Range query: request rate by service over 6 hours
cubeapm metrics query-range 'sum by (service) (rate(http_requests_total[5m]))' --last 6h --step 5m -o json

# Range query: error rate percentage over 24 hours
cubeapm metrics query-range 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100' --last 24h --step 15m -o json
```

### Explore metric metadata

```bash
# List all label names
cubeapm metrics labels --last 24h -o json

# List values for a specific label
cubeapm metrics label-values job -o json
cubeapm metrics label-values instance --last 24h -o json

# List all metric names
cubeapm metrics label-values __name__ -o json

# Find time series matching a selector
cubeapm metrics series --match 'http_requests_total' -o json

# Find series with regex matching
cubeapm metrics series --match '{__name__=~"http_.*"}' --last 24h -o json

# Limit returned series
cubeapm metrics series --match 'http_requests_total' --limit 50 -o json
```

### Query logs with LogsQL

```bash
# Search all logs in the last hour
cubeapm logs query '*' --last 1h -o json

# Search for errors in a specific service
cubeapm logs query 'error' --service api-gateway --last 30m -o json

# Filter by log level
cubeapm logs query '*' --service payments --level error --last 1h -o json

# Filter by stream
cubeapm logs query 'timeout' --stream '{host="web-1"}' --last 2h -o json

# Complex LogsQL expression
cubeapm logs query 'error AND service:api AND NOT health_check' --last 30m -o json

# Limit number of results (default 100)
cubeapm logs query 'status:500' --limit 50 -o json

# Show log volume over time (histogram)
cubeapm logs hits --query 'error' --last 24h --step 1h -o json

# Show volume for a specific service
cubeapm logs hits --query 'service:api-gateway' --last 6h --step 15m -o json
```

### Log aggregation and stats

```bash
# Count log entries by service
cubeapm logs stats 'error | stats count() by (service)' --last 24h -o json

# Count errors by log level
cubeapm logs stats '_time:1h | stats count() by (level)' -o json

# Count unique users per service
cubeapm logs stats '* | stats count_uniq(user_id) by (service)' --last 1h -o json

# Get top error messages
cubeapm logs stats 'level:error | stats count() by (_msg)' --last 1h -o json
```

### Explore log schema

```bash
# List all field names and their hit counts
cubeapm logs field-names --last 1h -o json

# List fields in error logs only
cubeapm logs field-names --query 'level:error' --last 1h -o json

# List values for a specific field
cubeapm logs field-values status --last 1h -o json
cubeapm logs field-values service --last 1h -o json
cubeapm logs field-values level --last 24h -o json

# List values filtered by a query
cubeapm logs field-values host --query 'error' --limit 50 -o json

# List log streams
cubeapm logs streams --last 1h -o json

# List streams for a specific service
cubeapm logs streams --query 'service:api-gateway' --last 1h -o json
```

### Manage log deletion

```bash
# Start a deletion task (uses admin port 3199)
cubeapm logs delete run '_time:<24h AND service:test'

# List active deletion tasks
cubeapm logs delete list -o json

# Stop a running deletion task
cubeapm logs delete stop <task-id>
```

### Ingest data

```bash
# Ingest metrics in Prometheus exposition format
cubeapm ingest metrics --format prometheus --file metrics.txt

# Pipe metrics from stdin
curl -s http://localhost:9090/metrics | cubeapm ingest metrics --format prometheus

# Ingest OTLP protobuf metrics
cubeapm ingest metrics --format otlp --file metrics.pb

# Ingest logs as JSON lines
cubeapm ingest logs --format jsonline --file logs.jsonl

# Pipe logs from stdin
cat logs.jsonl | cubeapm ingest logs --format jsonline

# Ingest OTLP protobuf logs
cubeapm ingest logs --format otlp --file logs.pb

# Ingest Loki-format logs
cubeapm ingest logs --format loki --file loki-push.json

# Ingest Elasticsearch bulk format logs
cubeapm ingest logs --format elastic --file elastic-bulk.ndjson
```

### Configuration management

```bash
# View full resolved configuration
cubeapm config view

# Set a configuration value
cubeapm config set server cubeapm.example.com

# Get a configuration value
cubeapm config get server

# List all profiles
cubeapm config profiles list

# Switch active profile
cubeapm config profiles use production

# Delete a profile
cubeapm config profiles delete staging
```

## LogsQL Quick Reference

LogsQL is the query language for logs (VictoriaLogs-compatible):

| Pattern | Description |
|---------|-------------|
| `error` | Keyword search in log message |
| `service:api-gateway` | Field filter (exact match) |
| `error AND service:api` | AND combination |
| `error OR warning` | OR combination |
| `NOT health_check` | Negation |
| `_stream:{host="web-1"}` | Stream filter |
| `re("pattern")` | Regex match |
| `_time:1h` | Time filter (inline) |
| `* \| stats count() by (service)` | Stats pipeline |
| `* \| stats count_uniq(user_id) by (service)` | Unique count aggregation |

## PromQL Quick Reference

PromQL is the query language for metrics (Prometheus-compatible):

| Pattern | Description |
|---------|-------------|
| `up` | Simple metric query |
| `http_requests_total{method="GET"}` | Label filtering |
| `rate(http_requests_total[5m])` | Rate of counter over 5 minutes |
| `sum by (service) (rate(requests_total[5m]))` | Aggregation by label |
| `http_requests_total{status=~"5.."}` | Regex label match |
| `histogram_quantile(0.99, rate(bucket[5m]))` | Histogram percentile |
| `rate(errors[5m]) / rate(requests[5m]) * 100` | Error rate percentage |

## Tips for Agents

- Always use `-o json` when you need to parse output programmatically.
- The typical trace investigation flow is: `services` -> `operations` -> `search` -> `get`.
- Use `--last` for relative time ranges (most common). It is simpler than `--from`/`--to`.
- Duration filter values use Go-style notation: `500ms`, `1s`, `100us`, `5m`.
- Trace IDs are 32-character hex strings. Get them from `traces search` output.
- The `traces get` command shows a visual waterfall view in table mode, but `-o json` gives full span data.
- For metrics, use `query` for current values and `query-range` for time series data.
- Range queries auto-calculate step if `--step` is omitted (~250 data points).
- The `--service`, `--level`, and `--stream` flags on `logs query` are convenience shortcuts that prepend filters to the LogsQL expression.
- Log stats queries must contain a `| stats` pipe (e.g., `'error | stats count() by (service)'`).
- Log deletion uses the admin port (default 3199), not the query port.
- Ingest commands use the ingest port (default 3130), not the query port.
- The `dependencies --dot` output can be piped to Graphviz tools (`dot`, `neato`) for visualization.
- Command aliases: `traces`/`trace`, `metrics`/`metric`, `logs`/`log`, `services`/`svc`, `operations`/`ops`, `dependencies`/`deps`, `query-range`/`range`, `field-names`/`fields`.

## Complete Command Reference

### Top-level commands

| Command | Description |
|---------|-------------|
| `cubeapm login` | Interactively configure a connection profile |
| `cubeapm version` | Print CLI version |
| `cubeapm update` | Check for and install CLI updates |

### `cubeapm config` -- Manage CLI configuration

| Command | Description |
|---------|-------------|
| `cubeapm config view` | Display the full resolved configuration |
| `cubeapm config set <key> <value>` | Set a configuration value |
| `cubeapm config get <key>` | Get a configuration value |
| `cubeapm config profiles list` | List all profiles (active profile marked with *) |
| `cubeapm config profiles use <name>` | Set the active profile |
| `cubeapm config profiles delete <name>` | Delete a profile |

### `cubeapm traces` (alias: `trace`) -- Distributed traces

| Command | Description |
|---------|-------------|
| `cubeapm traces services` | List all services (alias: `svc`) |
| `cubeapm traces operations <service>` | List operations for a service (alias: `ops`; --span-kind) |
| `cubeapm traces search` | Search traces (--service, --env, --query, --status, --min-duration, --max-duration, --tags, --span-kind, --limit, --last/--from/--to) |
| `cubeapm traces get <trace-id>` | Get a trace by ID (waterfall in table mode; --last/--from/--to) |
| `cubeapm traces dependencies` | Show service dependency graph (alias: `deps`; --dot, --last/--from/--to) |

### `cubeapm metrics` (alias: `metric`) -- Prometheus-compatible metrics

| Command | Description |
|---------|-------------|
| `cubeapm metrics query <promql>` | Execute an instant PromQL query (--time) |
| `cubeapm metrics query-range <promql>` | Execute a range PromQL query (alias: `range`; --step, --last/--from/--to) |
| `cubeapm metrics labels` | List all metric label names (--last/--from/--to) |
| `cubeapm metrics label-values <label>` | List values for a label (--last/--from/--to) |
| `cubeapm metrics series` | Find time series (--match, --limit, --last/--from/--to) |

### `cubeapm logs` (alias: `log`) -- VictoriaLogs-compatible logs

| Command | Description |
|---------|-------------|
| `cubeapm logs query <logsql>` | Query logs (--service, --level, --stream, --limit, --last/--from/--to) |
| `cubeapm logs hits` | Show log volume over time (--query, --step, --last/--from/--to) |
| `cubeapm logs stats <logsql>` | Execute a stats/aggregation query (--last/--from/--to) |
| `cubeapm logs streams` | List log streams and entry counts (--query, --last/--from/--to) |
| `cubeapm logs field-names` | List log field names and hit counts (alias: `fields`; --query, --last/--from/--to) |
| `cubeapm logs field-values <field>` | List values for a log field (--query, --limit, --last/--from/--to) |
| `cubeapm logs delete run <filter>` | Start a log deletion task |
| `cubeapm logs delete list` | List active deletion tasks |
| `cubeapm logs delete stop <task-id>` | Stop a running deletion task |

### `cubeapm ingest` -- Push data to CubeAPM

| Command | Description |
|---------|-------------|
| `cubeapm ingest metrics` | Push metrics data (--format: prometheus/otlp; --file) |
| `cubeapm ingest logs` | Push log data (--format: jsonline/otlp/loki/elastic; --file) |

## Global Flags

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: table (default), json, yaml |
| `--profile <name>` | Configuration profile to use |
| `--server <addr>` | CubeAPM server address override |
| `--token <token>` | Authentication token override |
| `--query-port <port>` | Query port override (default: 3140) |
| `--ingest-port <port>` | Ingest port override (default: 3130) |
| `--admin-port <port>` | Admin port override (default: 3199) |
| `--no-color` | Disable colored output |
| `--verbose` | Enable verbose HTTP request logging |
