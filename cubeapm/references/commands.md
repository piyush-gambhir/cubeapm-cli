# CubeAPM CLI Command Reference

Complete reference for all `cubeapm` commands, flags, and options.

---

## Global Flags

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
| `--verbose` | | bool | `false` | Enable verbose HTTP request logging (written to stdout) |
| `--read-only` | | bool | `false` | Block every write command (see [Safety settings](#safety-settings)) |
| `--no-input` | | bool | `false` | Disable all interactive prompts (`login` fails instead; `update` needs `--yes`) |
| `--quiet` | `-q` | bool | `false` | Suppress informational output |

---

## Top-Level Commands

### `cubeapm login`

Interactively configure a connection profile. Prompts for profile name, server address, authentication method (email/password or none), and port configuration. Tests the connection and saves the profile to `~/.config/cubeapm-cli/config.yaml`.

```bash
cubeapm login
```

### `cubeapm version`

Print CLI version, commit hash, and build date. When an earlier command already checked for releases, it also prints `latest:` and `update_available:` from the local cache (never from the network).

```bash
cubeapm version
```

### `cubeapm update`

Check for and install the latest release on macOS, Linux, and Windows.

```bash
cubeapm update --check           # Report current and latest versions
cubeapm update --check -o json   # The same, as JSON
cubeapm update                   # Ask, then install
cubeapm update --yes             # Install without asking
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--check` | | bool | `false` | Only check for updates, do not install |
| `--yes` | `-y` | bool | `false` | Install without asking for confirmation |

`update` downloads the release archive for your OS and architecture, verifies its SHA-256 checksum against the release's `checksums.txt`, and replaces the binary in place. On Windows the running `cubeapm.exe` is renamed to `cubeapm.exe.old` (deleted on a later run) and the new one takes its place. If the binary's directory is not writable, `update` fails and leaves the current binary untouched: re-run it with `sudo`, or reinstall with the install script into a writable directory (`INSTALL_DIR=~/.local/bin`). In a terminal it asks `Update now? [Y/n]`; under `--no-input`, or without a terminal, it needs `--yes`. A binary that `make install` put in a Go bin directory (`$GOBIN`, `$GOPATH/bin`, or `~/go/bin`) is not replaced: `update` prints `git pull && make install` (run it in your `cubeapm-cli/cli-go` checkout) instead. `update` does not work on a `dev` build.

`update --check` always asks GitHub (bypassing the daily cache) and exits 0 whether or not an update exists. With `-o json` it prints `current_version`, `latest_version`, `update_available`, `release_url`, and `install_method` (`self` or `go`). Read-only mode blocks `update` but allows `update --check`.

**Update notice.** In an interactive terminal, cubeapm checks GitHub for a new release at most once a day, in the background (cached in `update-check.json` in the config directory), and prints a notice on stderr after the command's output, at most once a day per release:

```
A new version of cubeapm is available: v<current> -> v<latest>
Update with: cubeapm update
Release notes: https://github.com/piyush-gambhir/cubeapm-cli/releases/tag/v<latest>
```

There is no check, and no network request, when stderr is not a terminal, when `CI` is set, with `--quiet` or `CUBEAPM_QUIET`, when `CUBEAPM_NO_UPDATE_NOTIFIER=1` or `NO_UPDATE_NOTIFIER=1` is set, on `dev` builds, or for `update`, `version`, `completion`, and `help`. Scripts, CI jobs, and coding agents therefore never see it.

---

## Traces Commands

Command group: `cubeapm traces` (alias: `cubeapm trace`)

Query and inspect distributed traces via the Jaeger-compatible API. All commands use the query port (default: 3140).

### `traces services`

List all services that have reported traces. If the Jaeger services endpoint is unavailable, the CLI falls back to the `service` metric label values.

```
cubeapm traces services [flags]
```

Alias: `cubeapm traces svc`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--env` | string | | List only services seen in this environment (metrics-derived from the `env` and `cube.environment` labels; a trace-only service may not appear). Sent unchanged, so use the deployment's value, for example `PROD` |
| `--from` | string | | Start time (only used with `--env`) |
| `--to` | string | | End time (only used with `--env`) |
| `--last` | string | | Relative duration from now (only used with `--env`; default window is the last 1 hour) |

JSON output: array of `{"SERVICE": "..."}`.

**Examples:**

```bash
cubeapm traces services
cubeapm traces services -o json
cubeapm traces services --env PROD --last 24h -o json
```

### `traces operations`

List all operations (endpoints/methods) for a service.

```
cubeapm traces operations <service> [flags]
```

Alias: `cubeapm traces ops`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--span-kind` | string | | Filter by span kind: `client`, `server`, `producer`, `consumer`, `internal` (empty lists all kinds) |

JSON output: array of `{"OPERATION": "...", "SPAN_KIND": "..."}`. Not every CubeAPM deployment exposes this endpoint; when it returns "unsupported path", use `traces search` and read the `OPERATION` column.

**Examples:**

```bash
cubeapm traces operations api-gateway
cubeapm traces operations api-gateway --span-kind server
cubeapm traces operations api-gateway -o json
```

### `traces search`

Search for traces matching the given criteria.

```
cubeapm traces search [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--service` | string | | Filter by service name |
| `--env` | string | | Filter by environment; sent to the server unchanged (case-sensitive). CubeAPM values are usually upper-case: `PROD`, `UAT` |
| `--query` | string | | Filter by operation name |
| `--status` | string | | Filter by span status: `error`, `ok` |
| `--min-duration` | string | | Minimum trace duration (e.g., `500ms`, `1s`) |
| `--max-duration` | string | | Maximum trace duration (e.g., `5s`, `10s`) |
| `--tags` | string[] | | Filter by span tag key=value (repeatable) |
| `--span-kind` | string | `server` | Filter by span kind: `client`, `server`, `producer`, `consumer`, `internal`. Defaults to `server` because some deployments require a value |
| `--index` | string | `cube:latency` | CubeAPM trace index to query (e.g. `cube:latency`, `cube:error`) |
| `--limit` | int | `20` | Maximum number of traces to return |
| `--from` | string | | Start time (RFC3339, Unix, or relative) |
| `--to` | string | | End time (RFC3339, Unix, or relative) |
| `--last` | string | | Relative duration from now (e.g., `1h`, `30m`) |

Many deployments require both `--service` and `--env`. JSON output: array of objects with `TRACE_ID`, `SERVICE`, `OPERATION`, `DURATION`, `STATUS`, `TIMESTAMP` (one per trace, from its root span).

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
cubeapm traces search --service payments --env PROD --span-kind server

# Search the error index
cubeapm traces search --index cube:error --service payments --env PROD --last 30m

# Search with a custom time range
cubeapm traces search --service auth --from 2024-01-15T00:00:00Z --to 2024-01-15T12:00:00Z

# Return more results
cubeapm traces search --service api-gateway --limit 100

# Output as JSON
cubeapm traces search --service api-gateway -o json
```

### `traces get`

Retrieve and display a specific trace by its trace ID.

```
cubeapm traces get <trace-id> [flags]
```

In table mode (default), renders a visual waterfall/tree view showing parent-child span relationships, service names, operations, durations, and status codes. In JSON/YAML mode, returns the full Jaeger-format trace data.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | | Start time of the lookup window |
| `--to` | string | | End time of the lookup window |
| `--last` | string | | Relative duration from now |

The lookup window is always sent; with no time flags it is the last 1 hour, so pass `--last` or `--from`/`--to` for older traces.

**Examples:**

```bash
# Get a trace (waterfall view)
cubeapm traces get abc123def456789

# Get a trace as JSON (full span data)
cubeapm traces get abc123def456789 -o json

# Narrow time range for faster lookup
cubeapm traces get abc123def456789 --from 2024-01-15T00:00:00Z --to 2024-01-15T12:00:00Z
```

### `traces dependencies`

Show the service dependency graph.

```
cubeapm traces dependencies [flags]
```

Alias: `cubeapm traces deps`

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

# Show dependencies for the last 24 hours as JSON
cubeapm traces dependencies --last 24h -o json

# Export as Graphviz DOT and render to PNG
cubeapm traces dependencies --last 24h --dot | dot -Tpng -o deps.png
```

`--dot` writes DOT regardless of `-o`. Otherwise JSON output is an array of `{"PARENT", "CHILD", "CALL_COUNT"}` objects.

### `traces callers`

Rank the services making outbound HTTP calls to a host, by call rate. Runs the instant PromQL query `topk(<topk>, sum by (service) (rate(cube_apm_latency_count{group_name="HTTP <host>",span_kind="client"}[<window>])))`.

```
cubeapm traces callers [flags]
```

At least one of `--host` or `--service` is required.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--host` | string | | Target host, matched against the `group_name` label (`HTTP ` is prepended if missing). If omitted, the host defaults to `api.spyne.ai` |
| `--service` | string | | Keep only client spans whose `span_name` contains this service name |
| `--window` | string | `2m` | PromQL rate window |
| `--topk` | int | `10` | Maximum number of callers to return |
| `--from` | string | | Accepted, but only the end of the resolved range is used |
| `--to` | string | | Evaluation time of the instant query |
| `--last` | string | | Evaluation time is now |

The query is evaluated at a single instant (the end of the time range, `now` by default), so the rate covers only `--window` before that instant. JSON output: array of `{"CALLER_SERVICE", "CALLS_PER_SEC"}` sorted by rate, highest first.

**Examples:**

```bash
cubeapm traces callers --host api.example.com --last 1h
cubeapm traces callers --host api.example.com --to 2026-04-19T13:50:00Z --window 5m --topk 20 -o json
cubeapm traces callers --host api.example.com --service MEDIA-SERVICE
```

---

## Metrics Commands

Command group: `cubeapm metrics` (alias: `cubeapm metric`)

Query Prometheus-compatible metrics using PromQL. All commands use the query port (default: 3140).

### `metrics query`

Execute an instant PromQL query at a single point in time.

```
cubeapm metrics query <promql> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--time` | string | now | Evaluation time (RFC3339, Unix, or relative like `now-1h`) |

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

### `metrics query-range`

Execute a range PromQL query over a time window. Returns a matrix of time series with multiple data points per series.

```
cubeapm metrics query-range <promql> [flags]
```

Alias: `cubeapm metrics range`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--step` | string | auto | Query resolution step (e.g., `15s`, `1m`, `5m`, `1h`). If omitted, the CLI uses the range divided by 250 (minimum `1s`) |
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

### `metrics labels`

List all available metric label names.

```
cubeapm metrics labels [flags]
```

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

### `metrics label-values`

List all values for a specific metric label.

```
cubeapm metrics label-values <label> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--match` | string[] | | Series selector that scopes the returned values (repeatable, ORed), e.g. `'{env="PROD"}'` |
| `--like` | string | | Case-insensitive substring filter applied to the returned values |
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

# Services in one environment, narrowed by substring
cubeapm metrics label-values service.name --match '{env="PROD"}' --like media

# Output as JSON
cubeapm metrics label-values job -o json
```

### `metrics series`

Find time series matching label selectors. At least one `--match` selector is required.

```
cubeapm metrics series [flags]
```

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

## Logs Commands

Command group: `cubeapm logs` (alias: `cubeapm log`)

Query and manage logs using LogsQL syntax (VictoriaLogs-compatible). Query commands use the query port (default: 3140). Deletion commands use the admin port (default: 3199).

### `logs query`

Query logs using LogsQL syntax. With `-o json` or `-o yaml`, entries stream one object at a time, not as an array: JSON is a sequence of objects, YAML is one document per entry separated by `---`.

```
cubeapm logs query <logsql> [flags]
```

The `--service`, `--level`, and `--stream` flags are convenience shortcuts that prepend filters to the LogsQL expression.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--service` | string | | Filter by service name (prepends `service:<value>` to query) |
| `--level` | string | | Filter by log level: `error`, `warn`, `info`, `debug` (prepends `level:<value>`) |
| `--stream` | string | | Filter by log stream (prepends `_stream:<value>`) |
| `--limit` | int | `100` | Maximum number of log entries to return |
| `--from` | string | | Start time |
| `--to` | string | | End time |
| `--last` | string | | Relative duration from now |

**LogsQL syntax:**

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

### `logs hits`

Show log volume over time (histogram of matching entries).

```
cubeapm logs hits [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--query` | string | `*` | LogsQL query to filter entries |
| `--step` | string | auto | Time bucket size (e.g., `5m`, `1h`). If omitted, the CLI uses the range divided by 60 (minimum `1s`) |
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

### `logs status`

Probe the effective log retention. Samples 24-hour hit buckets over the last `--lookback` days and reports the oldest and newest non-empty buckets.

```
cubeapm logs status [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--query` | string | `*` | LogsQL query to probe (e.g. `'service.name:api-gateway'`) |
| `--lookback` | int | `30` | Lookback window in days |

JSON output is one object: `query`, `lookbackDays`, `hasLogs`, `nonEmptyDays`, and when present `earliestNonZeroBucket`, `latestNonZeroBucket`, `retentionHours`, `note`.

**Examples:**

```bash
cubeapm logs status
cubeapm logs status --query 'service.name:api-gateway' --lookback 60 -o json
```

### `logs stats`

Execute a LogsQL stats/aggregation query. The query must contain a `| stats` pipe.

```
cubeapm logs stats <logsql> [flags]
```

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

### `logs streams`

List log streams and their entry counts.

```
cubeapm logs streams [flags]
```

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

### `logs field-names`

List all log field names and their hit counts.

```
cubeapm logs field-names [flags]
```

Alias: `cubeapm logs fields`

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

### `logs field-values`

List values for a specific log field.

```
cubeapm logs field-values <field> [flags]
```

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

### `logs delete run`

Start a log deletion task. **WARNING:** Deletion is irreversible. Preview with `cubeapm logs query` first.

Uses the admin API port (default: 3199).

```
cubeapm logs delete run <filter>
```

**Examples:**

```bash
cubeapm logs delete run '_time:<24h AND service:test'
cubeapm logs delete run '_stream:{env="staging"}'
cubeapm logs delete run '_time:<7d AND level:debug'
```

### `logs delete list`

List active deletion tasks. JSON output: array of `{"TASK_ID", "FILTER", "STATUS", "PROGRESS"}`. With no tasks it prints `No active deletion tasks.` instead. Allowed in read-only mode.

```
cubeapm logs delete list
```

Alias: `cubeapm logs delete ls`

### `logs delete stop`

Stop a running deletion task. Entries already deleted are not restored.

```
cubeapm logs delete stop <task-id> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--if-exists` | bool | `false` | Succeed silently if the task does not exist |

`logs delete run` and `logs delete stop` are blocked in read-only mode.

---

## Ingest Commands

Command group: `cubeapm ingest`

Push data to CubeAPM ingest endpoints. All ingest commands use the ingest port (default: 3130).

### `ingest metrics`

Push metrics data to CubeAPM.

```
cubeapm ingest metrics [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `prometheus` | Data format: `prometheus`, `otlp`, `remote-write` |
| `--file` | string | `-` (stdin) | File path or `-` for stdin |

**Format details:**

- **prometheus** -- Prometheus text exposition format (POST `/api/metrics/v1/save`)
- **otlp** -- OpenTelemetry metrics, protobuf binary (POST `/api/metrics/v1/save/otlp`)
- **remote-write** -- Prometheus remote write, Snappy-compressed protobuf (POST `/api/metrics/api/v1/write`)

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

# Prometheus remote write (Snappy-compressed protobuf)
cubeapm ingest metrics --format remote-write --file remote-write.pb
```

### `ingest logs`

Push log data to CubeAPM.

```
cubeapm ingest logs [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `jsonline` | Data format: `jsonline`, `otlp`, `loki`, `elastic` |
| `--file` | string | `-` (stdin) | File path or `-` for stdin |

**Format details:**

- **jsonline** -- Newline-delimited JSON. Each line: `{"_time":"...","_msg":"...","service":"..."}`
- **otlp** -- OpenTelemetry Protocol (protobuf binary)
- **loki** -- Loki push API format (JSON with streams/values arrays)
- **elastic** -- Elasticsearch bulk format (NDJSON with action/document pairs)

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

## Config Commands

Command group: `cubeapm config`

Manage CLI configuration and connection profiles. Configuration is stored at `~/.config/cubeapm-cli/config.yaml`.

### `config view`

Show the full resolved configuration in YAML format.

```
cubeapm config view
```

### `config set`

Set a configuration value in the current profile. Blocked in read-only mode.

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

### `config get`

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

### `config profiles list`

List all profiles. Active profile is marked with `*`.

```
cubeapm config profiles list
```

Alias: `cubeapm config profiles ls`

### `config profiles use`

Set the active profile. Blocked in read-only mode.

```
cubeapm config profiles use <profile>
```

**Examples:**

```bash
cubeapm config profiles use production
cubeapm config profiles use staging
```

### `config profiles delete`

Delete a profile. Blocked in read-only mode.

```
cubeapm config profiles delete <profile>
```

Alias: `cubeapm config profiles rm`

---

## Time Range Reference

These commands accept `--last`, `--from`, and `--to`: `traces search`, `get`, `services` (only with `--env`), `dependencies`, `callers` (end time only); `metrics query-range`, `labels`, `label-values`, `series`; `logs query`, `hits`, `stats`, `streams`, `field-names`, `field-values`. `metrics query` uses `--time`, `logs status` uses `--lookback <days>`, and the other commands take no time flags. There are multiple ways to specify them:

| Method | Flags | Examples |
|--------|-------|---------|
| Relative | `--last <duration>` | `--last 1h`, `--last 30m`, `--last 2d`, `--last 1d12h` |
| Absolute (RFC3339) | `--from`, `--to` | `--from 2024-01-15T10:00:00Z --to 2024-01-15T12:00:00Z` |
| Absolute (Unix) | `--from`, `--to` | `--from 1705312800 --to 1705356000` |
| Relative from/to | `--from`, `--to` | `--from -2h --to -1h` |
| Date only | `--from` | `--from 2024-01-15` (midnight local time) |

Timestamps without a zone (`2024-01-15T10:00:00`) are local time. Default: if no time flags are provided, the default is the last 1 hour. `--last` wins over `--from`/`--to`; only `--from` means `--to` is now; only `--to` means `--from` is one hour earlier.

## Command Aliases

| Full Name | Alias |
|-----------|-------|
| `traces` | `trace` |
| `metrics` | `metric` |
| `logs` | `log` |
| `services` | `svc` |
| `operations` | `ops` |
| `dependencies` | `deps` |
| `query-range` | `range` |
| `field-names` | `fields` |
| `delete list` | `delete ls` |
| `profiles list` | `profiles ls` |
| `profiles delete` | `profiles rm` |

## Ports Reference

| Port | Default | Env Var | Flag | Used By |
|------|---------|---------|------|---------|
| Query | 3140 | `CUBEAPM_QUERY_PORT` | `--query-port` | `traces`, `metrics`, `logs query/hits/status/stats/streams/field-names/field-values` |
| Ingest | 3130 | `CUBEAPM_INGEST_PORT` | `--ingest-port` | `ingest metrics`, `ingest logs` |
| Admin | 3199 | `CUBEAPM_ADMIN_PORT` | `--admin-port` | `logs delete run/list/stop` |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `CUBEAPM_SERVER` | CubeAPM server address |
| `CUBEAPM_EMAIL` | Login email |
| `CUBEAPM_PASSWORD` | Login password |
| `CUBEAPM_QUERY_PORT` | Query port (default: 3140) |
| `CUBEAPM_INGEST_PORT` | Ingest port (default: 3130) |
| `CUBEAPM_ADMIN_PORT` | Admin port (default: 3199) |
| `CUBEAPM_READ_ONLY` | A true Go boolean (`true`, `1`, ...) turns read-only mode on; `false`/`0` never turns off a profile's `read_only: true` |
| `CUBEAPM_NO_INPUT` | `1` or `true` disables interactive prompts |
| `CUBEAPM_QUIET` | `1` or `true` suppresses informational output and the update notice |
| `CUBEAPM_NO_UPDATE_NOTIFIER`, `NO_UPDATE_NOTIFIER` | Any non-empty value turns off the update notice and its GitHub release check |
| `CI` | Any non-empty value turns off the update notice and its GitHub release check |
| `XDG_CONFIG_HOME` | Relocates the config file to `$XDG_CONFIG_HOME/cubeapm-cli/config.yaml` |

## Safety settings

- **Read-only** is only ever added, never removed: it is on when the profile has `read_only: true`, when `CUBEAPM_READ_ONLY` is a true Go boolean, or when `--read-only` is passed. `CUBEAPM_READ_ONLY=false` and `--read-only=false` cannot turn off a profile's `read_only: true`. It blocks `ingest metrics`, `ingest logs`, `logs delete run`, `logs delete stop`, `config set`, `config profiles use`, `config profiles delete`, and `update` (`update --check` still runs); all queries, `logs delete list`, `config view`/`get`, and `config profiles list` still run. `login` is not covered; it is interactive, so `--no-input` blocks it. For the `config` write commands and `update`, the active profile's `read_only` applies even when `--profile` names another profile, because they change the active profile (or the binary).
- **No-input** is on when `--no-input` is passed or `CUBEAPM_NO_INPUT` is `1`/`true`; `--no-input=false` does not override the environment.
- **Quiet** follows `--quiet` when the flag is given (including `--quiet=false`), otherwise `CUBEAPM_QUIET`.
