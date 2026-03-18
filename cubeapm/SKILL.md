---
name: cubeapm
description: "Expert guide for using the cubeapm CLI to query distributed traces, metrics, and logs from CubeAPM. Use this skill whenever the user mentions CubeAPM, distributed tracing, trace search, span analysis, trace waterfall, service dependencies, PromQL metrics queries, Prometheus-compatible metrics, VictoriaLogs, LogsQL, log querying, log streaming, observability, APM, application performance monitoring, trace investigation, error analysis, latency analysis, service maps, or any CubeAPM operations. Also trigger when the user wants to search traces, query metrics with PromQL, query logs with LogsQL, view service dependencies, ingest telemetry data, investigate errors or performance issues, or analyze distributed system behavior from the command line."
---

# CubeAPM CLI Skill

## 1. Prerequisites & Setup

### Installation

```bash
# From source (Go 1.21+)
go install github.com/piyush-gambhir/cubeapm-cli@latest

# From the install script
curl -sSL https://raw.githubusercontent.com/piyush-gambhir/cubeapm-cli/main/install.sh | bash

# Build from source
git clone https://github.com/piyush-gambhir/cubeapm-cli.git
cd cubeapm-cli
make build
# Binary is at ./bin/cubeapm
```

### Authentication

```bash
# Interactive login (prompts for server, token, ports, profile name)
cubeapm login

# Or set environment variables for non-interactive/CI use
export CUBEAPM_SERVER=cubeapm.example.com
export CUBEAPM_TOKEN=your-api-token

# Or use CLI flags (override everything)
cubeapm --server cubeapm.example.com --token your-token traces services
```

**Config priority:** CLI flags > environment variables > profile config (`~/.config/cubeapm/config.yaml`).

### Multi-Port Architecture

CubeAPM uses three ports for different purposes. You almost always only need the query port.

| Port | Default | Env Var | Purpose |
|------|---------|---------|---------|
| Query | 3140 | `CUBEAPM_QUERY_PORT` | Read queries: traces, metrics, logs |
| Ingest | 3130 | `CUBEAPM_INGEST_PORT` | Write/push: metrics ingestion, log ingestion |
| Admin | 3199 | `CUBEAPM_ADMIN_PORT` | Administration: log deletion tasks |

Override via flags (`--query-port`, `--ingest-port`, `--admin-port`), env vars, or profile config.

---

## 2. Core Principles for Agents

1. **ALWAYS use `-o json` for programmatic parsing.** Table output is for humans; JSON output is for agents and scripts.
2. **Time ranges:** `--last 1h` is the quickest way to specify a time window. Also supports `--from`/`--to` with RFC3339 timestamps, Unix timestamps, or relative values (`-30m`, `-2d`). Default is last 1 hour if no time flags are provided.
3. **PromQL and LogsQL are passed as positional args, not flags:**
   ```bash
   cubeapm metrics query 'rate(http_requests_total[5m])'
   cubeapm logs query 'error AND service:api'
   ```
4. **Use single quotes around PromQL/LogsQL** to avoid shell escaping issues with curly braces, parentheses, and pipes.
5. **Multi-port awareness:** Query port (3140) for reads, ingest port (3130) for writes, admin port (3199) for deletion. Usually only the query port matters.
6. **Trace investigation follows a standard flow:** `services` -> `operations` -> `search` -> `get` (waterfall).
7. **Duration filter values** use Go-style notation: `500ms`, `1s`, `100us`, `5m`.
8. **Trace IDs** are 32-character hex strings. Get them from `traces search` output.
9. **Range queries** auto-calculate step (~250 data points) if `--step` is omitted.
10. **Command aliases exist** for convenience: `traces`/`trace`, `metrics`/`metric`, `logs`/`log`, `services`/`svc`, `operations`/`ops`, `dependencies`/`deps`, `query-range`/`range`, `field-names`/`fields`.

---

## 3. Common Workflows

### Workflow 1: Investigate an error end-to-end (traces + metrics + logs)

```bash
# 1. List services to find what is reporting
cubeapm traces services -o json

# 2. Search for error traces in the last hour
cubeapm traces search --service api-gateway --status error --last 1h -o json

# 3. Get full trace detail (waterfall view in table, full span data in JSON)
cubeapm traces get <trace-id> -o json

# 4. Check error rate metric via PromQL
cubeapm metrics query 'rate(http_requests_total{service="api-gateway",status=~"5.."}[5m])' -o json

# 5. Search related logs around the same time
cubeapm logs query 'error' --service api-gateway --last 1h -o json
```

### Workflow 2: Find slow endpoints

```bash
# Search for traces slower than 500ms
cubeapm traces search --service api-gateway --min-duration 500ms --last 1h -o json

# Check which operations are slow
cubeapm traces operations api-gateway --span-kind server -o json

# Check P99 latency via PromQL
cubeapm metrics query 'histogram_quantile(0.99, sum by (le) (rate(http_duration_seconds_bucket{service="api-gateway"}[5m])))' -o json

# Check latency trend over time
cubeapm metrics query-range 'histogram_quantile(0.99, sum by (le) (rate(http_duration_seconds_bucket{service="api-gateway"}[5m])))' --last 6h --step 5m -o json
```

### Workflow 3: PromQL instant and range queries

```bash
# Instant query (returns current value)
cubeapm metrics query 'up' -o json

# Request rate per service
cubeapm metrics query 'sum by (service) (rate(http_requests_total[5m]))' -o json

# Error rate as a percentage
cubeapm metrics query 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100' -o json

# Range query: request rate over the last hour, 1-minute resolution
cubeapm metrics query-range 'rate(http_requests_total[5m])' --last 1h --step 1m -o json

# Range query: error rate over 24 hours
cubeapm metrics query-range 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100' --last 24h --step 15m -o json

# Query at a specific time
cubeapm metrics query 'up' --time 2024-01-15T10:00:00Z -o json
```

### Workflow 4: Explore metric metadata

```bash
# List all label names
cubeapm metrics labels --last 24h -o json

# List all metric names
cubeapm metrics label-values __name__ -o json

# List values for a specific label
cubeapm metrics label-values job -o json
cubeapm metrics label-values instance --last 24h -o json

# Find time series matching a selector
cubeapm metrics series --match 'http_requests_total' -o json

# Find series with regex matching
cubeapm metrics series --match '{__name__=~"http_.*"}' --last 24h -o json

# Limit returned series
cubeapm metrics series --match 'http_requests_total' --limit 50 -o json
```

### Workflow 5: Search and analyze traces

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

# Get a specific trace by ID (JSON gives full span data)
cubeapm traces get abc123def456789 -o json

# List all operations for a service
cubeapm traces operations api-gateway -o json

# List only server-side operations
cubeapm traces operations api-gateway --span-kind server -o json
```

### Workflow 6: LogsQL queries

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

# Pipe to jq for field extraction
cubeapm logs query 'error' -o json | jq '.["_msg"]'
```

### Workflow 7: Log volume and histograms

```bash
# Show log volume over time (histogram)
cubeapm logs hits --query 'error' --last 24h --step 1h -o json

# Show volume for a specific service
cubeapm logs hits --query 'service:api-gateway' --last 6h --step 15m -o json

# All log volume in 5-minute buckets
cubeapm logs hits --query '*' --last 1h --step 5m -o json
```

### Workflow 8: Log aggregation and stats

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

### Workflow 9: Explore log schema

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

### Workflow 10: Service dependency graph

```bash
# Get dependencies as JSON
cubeapm traces dependencies --last 1h -o json

# Show dependencies for the last 24 hours
cubeapm traces dependencies --last 24h -o json

# Generate Graphviz DOT for visualization
cubeapm traces dependencies --last 24h --dot | dot -Tpng -o deps.png
```

### Workflow 11: Ingest metrics

```bash
# Push metrics in Prometheus exposition format from a file
cubeapm ingest metrics --format prometheus --file metrics.txt

# Pipe from a Prometheus exporter
curl -s http://localhost:9090/metrics | cubeapm ingest metrics --format prometheus

# Ingest OTLP protobuf metrics
cubeapm ingest metrics --format otlp --file metrics.pb
```

### Workflow 12: Ingest logs

```bash
# Ingest logs as JSON lines from a file
cubeapm ingest logs --format jsonline --file logs.jsonl

# Pipe logs from stdin
cat logs.jsonl | cubeapm ingest logs --format jsonline

# Ingest Loki-format logs
cubeapm ingest logs --format loki --file loki-push.json

# Ingest Elasticsearch bulk format logs
cubeapm ingest logs --format elastic --file elastic-bulk.ndjson

# Ingest OTLP protobuf logs
cubeapm ingest logs --format otlp --file logs.pb
```

### Workflow 13: Delete old logs

```bash
# Preview what would be deleted first
cubeapm logs query '_time:<24h AND service:test' --limit 10

# Start a deletion task (uses admin port 3199)
cubeapm logs delete run '_time:<24h AND service:test'

# Delete old debug logs
cubeapm logs delete run '_time:<7d AND level:debug'

# Delete logs for a staging environment
cubeapm logs delete run '_stream:{env="staging"}'

# List active deletion tasks
cubeapm logs delete list -o json

# Stop a running deletion task
cubeapm logs delete stop <task-id>
```

### Workflow 14: Configuration management

```bash
# View full resolved configuration
cubeapm config view

# Set a configuration value
cubeapm config set server cubeapm.example.com
cubeapm config set output json

# Get a configuration value
cubeapm config get server
cubeapm config get current_profile

# List all profiles (active profile marked with *)
cubeapm config profiles list

# Switch active profile
cubeapm config profiles use production

# Delete a profile
cubeapm config profiles delete staging
```

### Workflow 15: Monitor service health

```bash
# Check all services
cubeapm traces services -o json

# Check service dependencies
cubeapm traces dependencies --last 24h -o json

# Check which targets are up
cubeapm metrics query 'up' -o json

# Log volume trend over 24h
cubeapm logs hits --query '*' --last 24h --step 1h -o json

# Error volume trend
cubeapm logs hits --query 'level:error' --last 24h --step 1h -o json
```

### Workflow 16: Compare time windows (before/after a deploy)

```bash
# Error rate before deploy
cubeapm metrics query 'rate(http_requests_total{status=~"5.."}[5m])' --time 2024-01-15T09:00:00Z -o json

# Error rate after deploy
cubeapm metrics query 'rate(http_requests_total{status=~"5.."}[5m])' --time 2024-01-15T11:00:00Z -o json

# Traces around deploy time
cubeapm traces search --service api-gateway --status error --from 2024-01-15T09:30:00Z --to 2024-01-15T10:30:00Z -o json

# Logs around deploy time
cubeapm logs query 'error OR fatal' --from 2024-01-15T09:30:00Z --to 2024-01-15T10:30:00Z -o json
```

### Workflow 17: Trace-to-log correlation

```bash
# 1. Find an error trace
cubeapm traces search --service payments --status error --last 1h -o json

# 2. Get the full trace to see span timing
cubeapm traces get <trace-id> -o json

# 3. Search logs within the same time window as the trace
cubeapm logs query 'service:payments' --from <span-start-time> --to <span-end-time> -o json

# 4. Correlate by trace ID if logs contain trace_id field
cubeapm logs query 'trace_id:<trace-id>' --last 1h -o json
```

### Workflow 18: Script-friendly output with jq

```bash
# Extract trace IDs and durations
cubeapm traces search --service api-gateway --last 1h -o json | jq '.[] | {traceID, duration, operationName}'

# Get a list of service names
cubeapm traces services -o json | jq '.[].name'

# Extract metric values
cubeapm metrics query 'up' -o json | jq '.data.result[] | {metric: .metric.instance, value: .value[1]}'

# Count errors per service from logs
cubeapm logs stats 'level:error | stats count() by (service)' --last 1h -o json | jq '.[] | {service, count}'
```

---

## 4. Command Reference (Compact)

See [references/commands.md](references/commands.md) for the full command reference.

### Top-level

| Command | Description |
|---------|-------------|
| `cubeapm login` | Interactively configure a connection profile |
| `cubeapm version` | Print CLI version, commit hash, build date |
| `cubeapm update` | Check for and install CLI updates (`--check` for dry run) |

### Traces (`cubeapm traces` / `trace`)

| Command | Key Flags | Description |
|---------|-----------|-------------|
| `traces services` | | List all services (alias: `svc`) |
| `traces operations <svc>` | `--span-kind` | List operations for a service (alias: `ops`) |
| `traces search` | `--service`, `--status`, `--min-duration`, `--max-duration`, `--tags`, `--query`, `--env`, `--span-kind`, `--limit`, `--last`/`--from`/`--to` | Search traces |
| `traces get <id>` | `--last`/`--from`/`--to` | Get trace by ID (waterfall in table mode) |
| `traces dependencies` | `--dot`, `--last`/`--from`/`--to` | Service dependency graph (alias: `deps`) |

### Metrics (`cubeapm metrics` / `metric`)

| Command | Key Flags | Description |
|---------|-----------|-------------|
| `metrics query <promql>` | `--time` | Instant PromQL query |
| `metrics query-range <promql>` | `--step`, `--last`/`--from`/`--to` | Range PromQL query (alias: `range`) |
| `metrics labels` | `--last`/`--from`/`--to` | List all label names |
| `metrics label-values <label>` | `--last`/`--from`/`--to` | List values for a label |
| `metrics series` | `--match`, `--limit`, `--last`/`--from`/`--to` | Find time series |

### Logs (`cubeapm logs` / `log`)

| Command | Key Flags | Description |
|---------|-----------|-------------|
| `logs query <logsql>` | `--service`, `--level`, `--stream`, `--limit`, `--last`/`--from`/`--to` | Query logs |
| `logs hits` | `--query`, `--step`, `--last`/`--from`/`--to` | Log volume histogram |
| `logs stats <logsql>` | `--last`/`--from`/`--to` | Stats/aggregation query (must contain `\| stats`) |
| `logs streams` | `--query`, `--last`/`--from`/`--to` | List log streams |
| `logs field-names` | `--query`, `--last`/`--from`/`--to` | List field names (alias: `fields`) |
| `logs field-values <field>` | `--query`, `--limit`, `--last`/`--from`/`--to` | List field values |
| `logs delete run <filter>` | | Start log deletion (uses admin port) |
| `logs delete list` | | List active deletion tasks |
| `logs delete stop <id>` | | Stop a deletion task |

### Ingest (`cubeapm ingest`)

| Command | Key Flags | Description |
|---------|-----------|-------------|
| `ingest metrics` | `--format` (prometheus/otlp), `--file` | Push metrics data |
| `ingest logs` | `--format` (jsonline/otlp/loki/elastic), `--file` | Push log data |

### Config (`cubeapm config`)

| Command | Description |
|---------|-------------|
| `config view` | Show resolved configuration |
| `config set <key> <value>` | Set config value (keys: server, token, query_port, ingest_port, admin_port, output) |
| `config get <key>` | Get config value |
| `config profiles list` | List profiles |
| `config profiles use <name>` | Switch active profile |
| `config profiles delete <name>` | Delete a profile |

### Global Flags

| Flag | Description |
|------|-------------|
| `-o, --output` | Output format: `table` (default), `json`, `yaml` |
| `--profile` | Configuration profile to use |
| `--server` | Server address override |
| `--token` | Auth token override |
| `--query-port` | Query port override (default: 3140) |
| `--ingest-port` | Ingest port override (default: 3130) |
| `--admin-port` | Admin port override (default: 3199) |
| `--no-color` | Disable colored output |
| `--verbose` | Enable verbose HTTP request logging |

---

## 5. Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| Connection refused on port 3140 | Server might use non-default port | Check `--query-port` flag or `CUBEAPM_QUERY_PORT` env var. Verify server is running. |
| Connection refused on port 3130 | Ingest port mismatch | Check `--ingest-port` flag. Only ingest commands use this port. |
| Connection refused on port 3199 | Admin port mismatch | Check `--admin-port` flag. Only `logs delete` commands use this port. |
| Empty trace results | Time range too narrow or wrong service | Try `--last 24h` for a wider window. Run `traces services` to verify service name. |
| PromQL parse error | Shell is interpreting special characters | Use single quotes around the entire PromQL expression. Avoid double quotes. |
| LogsQL parse error | Pipe character interpreted by shell | Use single quotes: `'error \| stats count() by (service)'` |
| "dev" version | Binary built from source without version tags | Normal for local builds. Use `cubeapm update` to get a release build. |
| Auth error / 401 | Missing or invalid token | Run `cubeapm login` to reconfigure, or check `CUBEAPM_TOKEN` env var. |
| No data for metrics query | Wrong metric name or labels | Use `cubeapm metrics label-values __name__` to discover available metric names. |
| `logs stats` returns error | Missing stats pipe | The query must contain `\| stats`. Example: `'error \| stats count() by (service)'` |
| Log deletion not working | Using wrong port | `logs delete` commands use the admin port (3199), not the query port. |

---

## 6. Quick Reference: Query Languages

### PromQL (Prometheus-compatible)

| Pattern | Description |
|---------|-------------|
| `up` | Simple metric query |
| `http_requests_total{method="GET"}` | Label filtering |
| `rate(http_requests_total[5m])` | Rate of counter over 5 minutes |
| `sum by (service) (rate(requests_total[5m]))` | Aggregation by label |
| `http_requests_total{status=~"5.."}` | Regex label match |
| `histogram_quantile(0.99, rate(bucket[5m]))` | Histogram percentile |
| `rate(errors[5m]) / rate(requests[5m]) * 100` | Error rate percentage |

### LogsQL (VictoriaLogs-compatible)

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
| `status:>400` | Numeric comparison |
| `* \| stats count() by (service)` | Stats pipeline |
| `* \| stats count_uniq(user_id) by (service)` | Unique count aggregation |

### LogsQL Stats Functions

`count()`, `count_uniq(field)`, `sum(field)`, `avg(field)`, `min(field)`, `max(field)`, `median(field)`, `quantile(0.99, field)`, `values(field)`

---

## 7. References

- Full command reference: [references/commands.md](references/commands.md)
- CLI README: [README.md](../README.md)
- Agent guide: [CLAUDE.md](../CLAUDE.md)
- CubeAPM documentation: [https://cubeapm.com](https://cubeapm.com)
