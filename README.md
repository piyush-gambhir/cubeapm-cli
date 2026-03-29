# CubeAPM CLI

A command-line interface for querying traces, metrics, and logs from [CubeAPM](https://cubeapm.com). Supports Jaeger-compatible traces, Prometheus-compatible metrics (PromQL), and VictoriaLogs-compatible logs (LogsQL).

Designed to be used both interactively and programmatically by scripts and coding agents (LLMs).

[![Go Version](https://img.shields.io/github/go-mod/go-version/piyush-gambhir/cubeapm-cli)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/piyush-gambhir/cubeapm-cli)](https://github.com/piyush-gambhir/cubeapm-cli/releases)
[![License](https://img.shields.io/github/license/piyush-gambhir/cubeapm-cli)](LICENSE)
[![CI](https://github.com/piyush-gambhir/cubeapm-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/piyush-gambhir/cubeapm-cli/actions/workflows/ci.yml)

## Features

- Full API coverage — every CubeAPM API endpoint accessible from the command line
- Multiple output formats — table, JSON, YAML (`-o json`)
- Profile management — multiple instances with `--profile`
- Auto-update — checks for new versions, `cubeapm update` to self-update
- Agent-friendly — comprehensive help text, structured output for LLM coding agents
- Cross-platform — macOS, Linux, Windows (amd64 and arm64)

## Installation

### From source (Go 1.21+)

```bash
go install github.com/piyush-gambhir/cubeapm-cli@latest
```

### From the install script

```bash
curl -sSL https://raw.githubusercontent.com/piyush-gambhir/cubeapm-cli/main/install.sh | bash
```

### Build from source

```bash
git clone https://github.com/piyush-gambhir/cubeapm-cli.git
cd cubeapm-cli
make build
# Binary is at ./bin/cubeapm
```

## Quick Start

```bash
# Install
curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/cubeapm-cli/main/install.sh | sh

# Authenticate
cubeapm login

# Start using
cubeapm traces services
cubeapm traces search --service api-gateway --last 1h -o json
```

## Authentication

### Interactive login

```bash
cubeapm login
```

This prompts for server address, authentication method (email/password or none), and port configuration. It tests the connection and saves the profile.

### Environment variables

```bash
export CUBEAPM_SERVER=cubeapm.example.com
export CUBEAPM_EMAIL=user@example.com
export CUBEAPM_PASSWORD=your-password

# Custom ports (these are the defaults)
export CUBEAPM_QUERY_PORT=3140
export CUBEAPM_INGEST_PORT=3130
export CUBEAPM_ADMIN_PORT=3199
```

### CLI flags (override everything)

```bash
cubeapm --server cubeapm.example.com --email user@example.com --password secret traces services
```

### Configuration priority

Settings are resolved in this order (highest priority first):

1. CLI flags (`--server`, `--email`, `--password`, `--query-port`, etc.)
2. Environment variables (`CUBEAPM_SERVER`, `CUBEAPM_EMAIL`, `CUBEAPM_PASSWORD`, etc.)
3. Profile configuration (`~/.config/cubeapm-cli/config.yaml`)

## Time Ranges

All query commands support flexible time ranges via `--from`, `--to`, and `--last` flags:

```bash
# Relative duration from now (most common)
--last 1h                # last 1 hour
--last 30m               # last 30 minutes
--last 2d                # last 2 days
--last 1d12h             # last 1 day and 12 hours

# RFC3339 timestamps
--from 2024-01-15T10:00:00Z --to 2024-01-15T12:00:00Z

# Unix timestamps
--from 1705312800 --to 1705320000

# Relative from/to
--from -2h --to -1h      # between 2 hours ago and 1 hour ago

# Date only (midnight UTC)
--from 2024-01-15

# Default: if no time flags, defaults to last 1 hour
```

## Output Formats

All commands support three output formats via the `-o` / `--output` flag:

```bash
# Table format (default) - human-readable columns
cubeapm traces services

# JSON format - for scripting, piping to jq, or machine consumption
cubeapm traces services -o json

# YAML format - for configuration-style output
cubeapm traces services -o yaml
```

Set a default format:

```bash
cubeapm config set output json
```

## Quick Start

```bash
# 1. Configure connection
cubeapm login

# 2. Discover services
cubeapm traces services

# 3. Search traces
cubeapm traces search --service api-gateway --last 1h

# 4. View a specific trace
cubeapm traces get <trace-id>

# 5. Query metrics (PromQL)
cubeapm metrics query 'rate(http_requests_total[5m])'

# 6. Query logs (LogsQL)
cubeapm logs query 'service:api-gateway AND level:error' --last 30m
```

---

## Commands

### Global Flags

These flags apply to all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | `table` | Output format: `table`, `json`, `yaml` |
| `--server` | | string | | Override CubeAPM server address |
| `--email` | | string | | Override login email |
| `--password` | | string | | Override login password |
| `--profile` | | string | | Use a specific connection profile |
| `--query-port` | | int | `3140` | Override query API port |
| `--ingest-port` | | int | `3130` | Override ingest API port |
| `--admin-port` | | int | `3199` | Override admin API port |
| `--no-color` | | bool | `false` | Disable colored output |
| `--verbose` | | bool | `false` | Enable verbose HTTP request logging |

---

### Traces

Query and inspect distributed traces via the Jaeger-compatible API.

#### `traces search`

Search for traces matching the given criteria.

```
cubeapm traces search [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--service` | string | | Filter by service name |
| `--env` | string | | Filter by environment tag |
| `--query` | string | | Filter by operation name |
| `--status` | string | | Filter by span status: `error`, `ok` |
| `--min-duration` | string | | Minimum trace duration (e.g., `500ms`, `1s`) |
| `--max-duration` | string | | Maximum trace duration (e.g., `5s`, `10s`) |
| `--tags` | string[] | | Filter by span tag key=value (repeatable) |
| `--span-kind` | string | | Filter by span kind: `client`, `server`, `producer`, `consumer`, `internal` |
| `--limit` | int | `20` | Maximum number of traces to return |
| `--from` | string | | Start time (RFC3339, Unix, or relative) |
| `--to` | string | | End time (RFC3339, Unix, or relative) |
| `--last` | string | | Relative duration from now (e.g., `1h`, `30m`) |

**Examples:**

```bash
# Search traces for a service in the last hour
cubeapm traces search --service api-gateway --last 1h

# Find slow traces (>500ms) with errors
cubeapm traces search --service payments --min-duration 500ms --status error

# Filter by operation name
cubeapm traces search --service api-gateway --query "GET /api/users" --last 1h

# Filter by span tags
cubeapm traces search --service api-gateway --tags "http.method=POST" --tags "http.status_code=500"

# Filter by environment and span kind
cubeapm traces search --service payments --env production --span-kind server

# Search with a custom time range
cubeapm traces search --service auth --from 2024-01-15T00:00:00Z --to 2024-01-15T12:00:00Z

# Return more results
cubeapm traces search --service api-gateway --limit 100

# Output as JSON
cubeapm traces search --service api-gateway -o json
```

#### `traces get`

Retrieve and display a specific trace by its trace ID.

```
cubeapm traces get <trace-id> [flags]
```

In table mode (default), renders a visual waterfall/tree view showing parent-child span relationships, service names, operations, durations, and status codes. In JSON/YAML mode, returns the full Jaeger-format trace data.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | | Start time (narrows lookup window) |
| `--to` | string | | End time (narrows lookup window) |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
# Get a trace (waterfall view)
cubeapm traces get abc123def456789

# Get a trace as JSON
cubeapm traces get abc123def456789 -o json

# Narrow time range for faster lookup
cubeapm traces get abc123def456789 --from 2024-01-15T00:00:00Z --to 2024-01-15T12:00:00Z
```

#### `traces services`

List all services that have reported traces.

```
cubeapm traces services
```

Alias: `cubeapm traces svc`

**Examples:**

```bash
cubeapm traces services
cubeapm traces services -o json
```

#### `traces operations`

List all operations (endpoints/methods) for a service.

```
cubeapm traces operations <service> [flags]
```

Alias: `cubeapm traces ops`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--span-kind` | string | | Filter by span kind: `client`, `server`, etc. |

**Examples:**

```bash
# List all operations for a service
cubeapm traces operations api-gateway

# List only server-side operations
cubeapm traces operations api-gateway --span-kind server

# Output as JSON
cubeapm traces operations api-gateway -o json
```

#### `traces dependencies`

Show the service dependency graph.

```
cubeapm traces dependencies [flags]
```

Alias: `cubeapm traces deps`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dot` | bool | `false` | Output in Graphviz DOT format |
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
# Show dependencies for the last hour
cubeapm traces dependencies

# Show dependencies for the last 24 hours
cubeapm traces dependencies --last 24h

# Export as Graphviz DOT and render to PNG
cubeapm traces dependencies --last 24h --dot | dot -Tpng -o deps.png

# Output as JSON
cubeapm traces dependencies --last 24h -o json
```

---

### Metrics

Query Prometheus-compatible metrics using PromQL.

#### `metrics query`

Execute an instant PromQL query at a single point in time.

```
cubeapm metrics query <promql> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--time` | string | now | Evaluation time (RFC3339, Unix, or relative like `now-1h`) |

**PromQL syntax reference:**

```
up                                          # Simple metric
rate(http_requests_total[5m])              # Rate of a counter
sum by (service) (rate(requests_total[5m]))# Aggregation
http_requests_total{method="GET"}          # Label filter
histogram_quantile(0.99, rate(http_duration_seconds_bucket[5m]))  # Percentile
```

**Examples:**

```bash
# Check which targets are up
cubeapm metrics query 'up'

# Request rate per service
cubeapm metrics query 'sum by (service) (rate(http_requests_total[5m]))'

# Query at a specific time
cubeapm metrics query 'up' --time now-1h

# Error rate as a percentage
cubeapm metrics query 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100'

# P99 latency
cubeapm metrics query 'histogram_quantile(0.99, sum by (le) (rate(http_duration_seconds_bucket[5m])))'

# Output as JSON
cubeapm metrics query 'up' -o json
```

#### `metrics query-range`

Execute a range PromQL query over a time window.

```
cubeapm metrics query-range <promql> [flags]
```

Alias: `cubeapm metrics range`

Returns a matrix of time series with multiple data points per series.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--step` | string | auto | Query resolution step (e.g., `15s`, `1m`, `5m`, `1h`). Auto-calculated if omitted (~250 data points). |
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
# Request rate over the last hour, 1-minute resolution
cubeapm metrics query-range 'rate(http_requests_total[5m])' --last 1h --step 1m

# Rate by service over 6 hours
cubeapm metrics query-range 'sum by (service) (rate(http_requests_total[5m]))' --last 6h --step 5m

# Error rate over the last day
cubeapm metrics query-range 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100' --last 24h --step 15m

# Specific time window
cubeapm metrics query-range 'up' --from 2024-01-15T00:00:00Z --to 2024-01-16T00:00:00Z --step 1h

# Output as JSON for graphing
cubeapm metrics query-range 'rate(http_requests_total[5m])' --last 1h -o json
```

#### `metrics labels`

List all available metric label names.

```
cubeapm metrics labels [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
cubeapm metrics labels
cubeapm metrics labels --last 24h
cubeapm metrics labels -o json
```

#### `metrics label-values`

List all values for a specific metric label.

```
cubeapm metrics label-values <label> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
# List all job names
cubeapm metrics label-values job

# List all metric names
cubeapm metrics label-values __name__

# List instances seen in the last 24 hours
cubeapm metrics label-values instance --last 24h

# Output as JSON
cubeapm metrics label-values job -o json
```

#### `metrics series`

Find time series matching label selectors.

```
cubeapm metrics series [flags]
```

At least one `--match` selector is required.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--match` | string[] | | Series selector (repeatable, PromQL label matcher syntax) |
| `--limit` | int | `0` | Maximum number of series to return (0 = unlimited) |
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
# Find series matching a metric name
cubeapm metrics series --match 'up{job="api"}'

# Find multiple metrics
cubeapm metrics series --match 'http_requests_total' --match 'process_cpu_seconds_total'

# Limit results
cubeapm metrics series --match 'http_requests_total' --limit 50

# Regex matching
cubeapm metrics series --match '{__name__=~"http_.*"}' --last 24h

# Output as JSON
cubeapm metrics series --match 'up' -o json
```

---

### Logs

Query and manage logs using LogsQL syntax (VictoriaLogs-compatible).

#### `logs query`

Query logs using LogsQL syntax.

```
cubeapm logs query <logsql> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--service` | string | | Filter by service name (prepends `service:<value>` to query) |
| `--level` | string | | Filter by log level: `error`, `warn`, `info`, `debug` (prepends `level:<value>`) |
| `--stream` | string | | Filter by log stream (prepends `_stream:<value>`) |
| `--limit` | int | `100` | Maximum number of log entries to return |
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**LogsQL syntax reference:**

```
*                          # Match all logs
error                      # Keyword search
service:api-gateway        # Field filter
error AND service:api      # Boolean AND
error OR warning           # Boolean OR
NOT health_check           # Boolean NOT
_stream:{host="web-1"}     # Stream filter
re("pattern.*")            # Regex match
_time:1h                   # Time filter within query
status:>400                # Numeric comparison
```

**Examples:**

```bash
# Search all logs in the last hour
cubeapm logs query '*'

# Search for errors in a specific service
cubeapm logs query 'error' --service api-gateway --last 30m

# Filter by log level
cubeapm logs query '*' --service payments --level error --last 1h

# Filter by stream
cubeapm logs query 'timeout' --stream '{host="web-1"}' --last 2h

# Complex LogsQL expression
cubeapm logs query 'error AND service:api AND NOT health_check' --last 30m

# Limit results
cubeapm logs query 'status:500' --limit 50

# Output as JSON for scripting
cubeapm logs query 'error' --service api-gateway -o json

# Pipe to jq
cubeapm logs query 'error' -o json | jq '.["_msg"]'

# Explicit time range
cubeapm logs query 'error' --from 2024-01-15T00:00:00Z --to 2024-01-15T12:00:00Z
```

#### `logs hits`

Show log volume over time (histogram of matching entries).

```
cubeapm logs hits [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--query` | string | `*` | LogsQL query to filter entries |
| `--step` | string | auto | Time bucket size (e.g., `5m`, `1h`) |
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
# All log volume over the last hour in 5-minute buckets
cubeapm logs hits --query '*' --last 1h --step 5m

# Error volume over the last 24 hours
cubeapm logs hits --query 'error' --last 24h --step 1h

# Volume for a specific service
cubeapm logs hits --query 'service:api-gateway' --last 6h --step 15m

# Output as JSON
cubeapm logs hits --query 'error' --last 24h --step 1h -o json
```

#### `logs stats`

Execute a LogsQL stats/aggregation query.

```
cubeapm logs stats <logsql> [flags]
```

The query must contain a `| stats` pipe.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Stats functions:** `count()`, `count_uniq(field)`, `sum(field)`, `avg(field)`, `min(field)`, `max(field)`, `median(field)`, `quantile(0.99, field)`, `values(field)`

**Examples:**

```bash
# Count entries by status
cubeapm logs stats '_time:1h | stats count() by (status)'

# Count errors by service
cubeapm logs stats 'error | stats count() by (service)' --last 24h

# Count unique users per service
cubeapm logs stats '* | stats count_uniq(user_id) by (service)' --last 1h

# Top error messages
cubeapm logs stats 'level:error | stats count() by (_msg)' --last 1h

# Output as JSON
cubeapm logs stats 'error | stats count() by (service)' --last 24h -o json
```

#### `logs streams`

List log streams and their entry counts.

```
cubeapm logs streams [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--query` | string | | LogsQL query to filter streams |
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
cubeapm logs streams --last 1h
cubeapm logs streams --query 'error' --last 24h
cubeapm logs streams --query 'service:api-gateway' --last 1h
cubeapm logs streams --last 1h -o json
```

#### `logs field-names`

List all log field names and their hit counts.

```
cubeapm logs field-names [flags]
```

Alias: `cubeapm logs fields`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--query` | string | | LogsQL query to filter which entries to inspect |
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
cubeapm logs field-names --last 1h
cubeapm logs field-names --query 'service:api' --last 24h
cubeapm logs field-names --query 'level:error' --last 1h
cubeapm logs fields --last 1h
cubeapm logs field-names --last 1h -o json
```

#### `logs field-values`

List values for a specific log field.

```
cubeapm logs field-values <field> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--query` | string | | LogsQL query to filter which entries to inspect |
| `--limit` | int | `100` | Maximum number of values to return |
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**Examples:**

```bash
cubeapm logs field-values status --last 1h
cubeapm logs field-values host --query 'error' --limit 50
cubeapm logs field-values level --last 24h
cubeapm logs field-values service --last 1h
cubeapm logs field-values status --last 1h -o json
```

#### `logs delete`

Manage log deletion tasks.

##### `logs delete run`

Start a log deletion task.

```
cubeapm logs delete run <filter>
```

**WARNING:** Deletion is irreversible. Preview with `cubeapm logs query` first.

Uses the admin API port (default: 3199).

**Examples:**

```bash
cubeapm logs delete run '_time:<24h AND service:test'
cubeapm logs delete run '_stream:{env="staging"}'
cubeapm logs delete run '_time:<7d AND level:debug'
```

##### `logs delete list`

List active deletion tasks.

```
cubeapm logs delete list
```

Alias: `cubeapm logs delete ls`

##### `logs delete stop`

Stop a running deletion task.

```
cubeapm logs delete stop <task-id>
```

---

### Ingest

Push data to CubeAPM ingest endpoints (default port: 3130).

#### `ingest metrics`

Push metrics data to CubeAPM.

```
cubeapm ingest metrics [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `prometheus` | Data format: `prometheus`, `otlp` |
| `--file` | string | `-` (stdin) | File path or `-` for stdin |

**Examples:**

```bash
# From a file
cubeapm ingest metrics --format prometheus --file metrics.txt

# From stdin
cat metrics.txt | cubeapm ingest metrics --format prometheus

# From a Prometheus exporter
curl -s http://localhost:9090/metrics | cubeapm ingest metrics --format prometheus

# OTLP protobuf
cubeapm ingest metrics --format otlp --file metrics.pb
```

#### `ingest logs`

Push log data to CubeAPM.

```
cubeapm ingest logs [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `jsonline` | Data format: `jsonline`, `otlp`, `loki`, `elastic` |
| `--file` | string | `-` (stdin) | File path or `-` for stdin |

**Format details:**

- **jsonline** - Newline-delimited JSON. Each line: `{"_time":"...","_msg":"...","service":"..."}`
- **otlp** - OpenTelemetry Protocol (protobuf binary)
- **loki** - Loki push API format (JSON with streams/values arrays)
- **elastic** - Elasticsearch bulk format (NDJSON with action/document pairs)

**Examples:**

```bash
# JSON line logs from a file
cubeapm ingest logs --format jsonline --file logs.jsonl

# From stdin
cat logs.jsonl | cubeapm ingest logs --format jsonline

# Loki format
cubeapm ingest logs --format loki --file loki-push.json

# Elasticsearch bulk format
cubeapm ingest logs --format elastic --file elastic-bulk.ndjson
```

---

### Config

Manage CLI configuration and connection profiles.

Configuration is stored at `~/.config/cubeapm/config.yaml`.

#### `config view`

Show the full resolved configuration in YAML format.

```
cubeapm config view
```

#### `config set`

Set a configuration value in the current profile.

```
cubeapm config set <key> <value>
```

**Valid keys:** `server`, `email`, `password`, `auth_method`, `query_port`, `ingest_port`, `admin_port`, `output`

**Examples:**

```bash
cubeapm config set server cubeapm.example.com
cubeapm config set output json
cubeapm config set query_port 3140
cubeapm config set email user@example.com
```

#### `config get`

Get a configuration value from the current profile.

```
cubeapm config get <key>
```

**Valid keys:** `server`, `email`, `password`, `auth_method`, `query_port`, `ingest_port`, `admin_port`, `output`, `current_profile`

**Examples:**

```bash
cubeapm config get server
cubeapm config get output
cubeapm config get current_profile
```

#### `config profiles list`

List all profiles. Active profile is marked with `*`.

```
cubeapm config profiles list
```

Alias: `cubeapm config profiles ls`

#### `config profiles use`

Set the active profile.

```
cubeapm config profiles use <profile>
```

**Examples:**

```bash
cubeapm config profiles use production
cubeapm config profiles use staging
```

#### `config profiles delete`

Delete a profile.

```
cubeapm config profiles delete <profile>
```

Alias: `cubeapm config profiles rm`

---

### Other Commands

#### `login`

Interactively configure a connection profile.

```
cubeapm login
```

Prompts for profile name, server address, authentication method (email/password or none), and port configuration. Tests the connection and saves the profile.

#### `version`

Print CLI version, commit hash, and build date.

```
cubeapm version
```

#### `update`

Check for and install the latest version.

```
cubeapm update           # Check and install
cubeapm update --check   # Only check, do not install
```

---

## Common Workflows

### Investigate an error in production

```bash
# 1. Find recent error traces for a service
cubeapm traces search --service payments --status error --last 1h

# 2. Get the full trace for a specific error
cubeapm traces get <trace-id-from-above>

# 3. Check error logs around the same time
cubeapm logs query 'level:error AND service:payments' --last 1h

# 4. Check error rate metric
cubeapm metrics query 'rate(http_requests_total{service="payments",status=~"5.."}[5m])'
```

### Find slow endpoints

```bash
# 1. Search for slow traces (>1s)
cubeapm traces search --service api-gateway --min-duration 1s --last 6h

# 2. Check which operations are slow
cubeapm traces operations api-gateway --span-kind server

# 3. Get P99 latency from metrics
cubeapm metrics query 'histogram_quantile(0.99, sum by (le) (rate(http_duration_seconds_bucket{service="api-gateway"}[5m])))'
```

### Explore log schema

```bash
# 1. Discover what fields are available
cubeapm logs field-names --last 1h

# 2. See what values a field has
cubeapm logs field-values service --last 1h
cubeapm logs field-values level --last 1h

# 3. Check log streams
cubeapm logs streams --last 1h

# 4. Run targeted queries
cubeapm logs query 'service:api-gateway AND level:error' --last 1h
```

### Monitor service health

```bash
# Check all services
cubeapm traces services

# Check service dependencies
cubeapm traces dependencies --last 24h

# Check upness
cubeapm metrics query 'up'

# Log volume trend
cubeapm logs hits --query '*' --last 24h --step 1h
```

## Ports Reference

| Port | Default | Environment Variable | Description |
|------|---------|---------------------|-------------|
| Query | 3140 | `CUBEAPM_QUERY_PORT` | Traces, metrics, and log queries |
| Ingest | 3130 | `CUBEAPM_INGEST_PORT` | Metrics and log data ingestion |
| Admin | 3199 | `CUBEAPM_ADMIN_PORT` | Administrative operations (log deletion) |

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Agent Skills

This CLI ships with an agent skill for coding agents (Claude, Cursor, Copilot, etc.):

```bash
npx skills add piyush-gambhir/cubeapm-cli@cubeapm
```

Once installed, coding agents automatically know how to use this CLI effectively.

## License

See [LICENSE](LICENSE) for details.
