# CubeAPM CLI - Authentication & Credentials Guide

This guide covers every aspect of authenticating and configuring the CubeAPM CLI to connect to your CubeAPM server. Whether you are running a quick local setup, automating in CI/CD, or managing multiple production instances, this document has you covered.

---

## Quick Start

### Option 1: Interactive Login (Recommended)

The fastest way to get started is the interactive `login` command:

```bash
cubeapm login
```

The CLI will walk you through each setting with sensible defaults:

```
Profile name [default]:
CubeAPM server: cubeapm.internal.example.com
API token (leave empty for no auth):
Query port [3140]:
Ingest port [3130]:
Admin port [3199]:

Testing connection... OK (12 services found)

Profile "default" saved (active)
Config written to /home/user/.config/cubeapm-cli/config.yaml
```

After login, all subsequent commands use the saved profile automatically:

```bash
cubeapm traces services
cubeapm traces search --service api-gateway --last 1h
cubeapm metrics query 'up'
cubeapm logs query 'error' --last 30m
```

### Option 2: Environment Variables

For non-interactive use (CI pipelines, scripts, coding agents), set environment variables instead:

```bash
export CUBEAPM_SERVER=cubeapm.example.com
export CUBEAPM_TOKEN=your-api-token

# Optional: override default ports
export CUBEAPM_QUERY_PORT=3140
export CUBEAPM_INGEST_PORT=3130
export CUBEAPM_ADMIN_PORT=3199

# Now run commands directly
cubeapm traces services -o json
```

### Option 3: CLI Flags (One-Off Commands)

Pass credentials directly on the command line for one-off use. Flags override everything else:

```bash
cubeapm --server cubeapm.example.com --token my-token traces services
```

---

## Getting Your Credentials

### Step 1: Find Your CubeAPM Server Address

The server address is the hostname or IP where your CubeAPM instance is running. Provide the hostname only -- the CLI adds the `http://` scheme automatically.

**Self-hosted (single binary):**

If you installed CubeAPM directly on a server, use that server's hostname or IP:

```bash
# Hostname
cubeapm login
# CubeAPM server: cubeapm.internal.example.com

# IP address
# CubeAPM server: 10.0.1.50
```

**Docker deployment:**

If CubeAPM runs in Docker on the same machine, use `localhost`. If it runs on a remote Docker host, use that host's address:

```bash
# Same machine
# CubeAPM server: localhost

# Remote Docker host
# CubeAPM server: docker-host.example.com
```

**Kubernetes deployment:**

Use the Kubernetes Service name (if running in-cluster) or set up port-forwarding for local access:

```bash
# In-cluster (from another pod)
# CubeAPM server: cubeapm.monitoring.svc.cluster.local

# Local access via port-forward (see Kubernetes Deployment section below)
# CubeAPM server: localhost
```

> **Note:** Provide the hostname only, not a full URL. The CLI constructs URLs internally by combining the scheme, hostname, and the appropriate port for each operation. If you include a scheme (e.g., `https://cubeapm.example.com`), the CLI will parse and use it -- but in most cases, just the hostname is sufficient.

### Step 2: Understand the Port Architecture

CubeAPM exposes multiple ports, each serving a different purpose. This is important because the CLI connects to three of them depending on the operation:

| Port | Purpose | Default | Used For |
|------|---------|---------|----------|
| 3125 | Web UI | 3125 | Browser access to the CubeAPM dashboard |
| 3130 | Ingest | 3130 | Receiving traces, metrics, and logs from agents |
| 3140 | Query API | 3140 | Querying data (traces, metrics, logs) |
| 3199 | Admin | 3199 | Administrative operations (log deletion) |
| 4317 | OTLP gRPC | 4317 | OpenTelemetry gRPC ingestion |
| 4318 | OTLP HTTP | 4318 | OpenTelemetry HTTP ingestion |

**Which ports does the CLI use?**

The CLI connects to three ports depending on the command:

| CLI Operation | Port Used | Default | Example Commands |
|---------------|-----------|---------|------------------|
| Querying traces, metrics, logs | Query port | 3140 | `traces search`, `metrics query`, `logs query` |
| Pushing data | Ingest port | 3130 | `ingest metrics`, `ingest logs` |
| Admin operations | Admin port | 3199 | `logs delete run`, `logs delete list` |

If your CubeAPM instance uses the default ports, you do not need to configure them -- the CLI uses the defaults automatically.

### Step 3: Get an API Token (Optional)

CubeAPM authentication is optional. Many internal deployments run without authentication enabled.

**When a token is needed:**

- Your CubeAPM server has authentication enabled in its configuration
- You are accessing CubeAPM over the public internet
- Your organization's security policy requires it

**When a token is NOT needed:**

- CubeAPM is deployed on an internal network with no auth configured
- You are running CubeAPM locally for development
- The CubeAPM server does not enforce token validation

**How to configure the token on the server side:**

CubeAPM supports token-based authentication configured in the server's settings. Consult your CubeAPM server documentation or your infrastructure team for the specific token value.

**Using the token with the CLI:**

```bash
# During interactive login
cubeapm login
# API token (leave empty for no auth): your-api-token-here

# Via environment variable
export CUBEAPM_TOKEN=your-api-token-here

# Via CLI flag
cubeapm --token your-api-token-here traces services

# Via config file
cubeapm config set token your-api-token-here
```

### Step 4: Use with the CLI

Once you have your server address, ports, and (optionally) a token, you can authenticate using any of the three methods:

**Interactive login (saves to config file):**

```bash
cubeapm login
```

**Environment variables (for scripts and CI):**

```bash
export CUBEAPM_SERVER=cubeapm.example.com
export CUBEAPM_TOKEN=your-api-token
cubeapm traces services
```

**CLI flags (for one-off commands):**

```bash
cubeapm --server cubeapm.example.com --token your-api-token traces services
```

**Config file set (programmatic, non-interactive):**

```bash
cubeapm config set server cubeapm.example.com
cubeapm config set token your-api-token
cubeapm config set query_port 3140
cubeapm config set ingest_port 3130
cubeapm config set admin_port 3199
```

---

## Understanding CubeAPM Authentication

### Token-Based Auth

When CubeAPM has authentication enabled, the CLI sends a Bearer token in the `Authorization` HTTP header with every request:

```
Authorization: Bearer your-api-token
```

Key behaviors:

- The token is sent on all requests to the query, ingest, and admin ports
- If the token is invalid or expired, the server returns HTTP 401 and the CLI displays: `authentication failed (HTTP 401): check your token with 'cubeapm login'`
- If the token lacks permissions for a specific operation, the server returns HTTP 403 and the CLI displays: `access denied (HTTP 403): your token may not have sufficient permissions`
- When `--verbose` mode is enabled, the CLI logs request and response headers but redacts the `Authorization` header value to `[REDACTED]` for security

### No Authentication

CubeAPM can run without any authentication. In this case, simply leave the token empty:

```bash
# Interactive login
cubeapm login
# API token (leave empty for no auth): <press Enter>

# Environment variable (just don't set CUBEAPM_TOKEN)
export CUBEAPM_SERVER=cubeapm.example.com
cubeapm traces services

# Config file
cubeapm config set server cubeapm.example.com
cubeapm config set token ""
```

**When running without authentication is acceptable:**

- Development and testing environments
- Internal networks with network-level access controls (firewalls, VPNs)
- CubeAPM instances behind an authenticating reverse proxy

**Security implications:**

- Anyone with network access to the CubeAPM ports can read all observability data
- Anyone with access to the admin port (3199) can delete logs
- Anyone with access to the ingest port (3130) can push arbitrary data

---

## Configuration

### Configuration Priority

Settings are resolved in this order, with the highest-priority source winning:

```
1. CLI flags          (--server, --token, --query-port, etc.)    ← highest
2. Environment vars   (CUBEAPM_SERVER, CUBEAPM_TOKEN, etc.)
3. Profile config     (~/.config/cubeapm-cli/config.yaml)        ← lowest
```

This means you can set a baseline in your config file, override per-environment with env vars, and override for a single command with flags.

### Config File

The CLI stores configuration at `~/.config/cubeapm-cli/config.yaml` (or `$XDG_CONFIG_HOME/cubeapm-cli/config.yaml` if `XDG_CONFIG_HOME` is set).

**Full config file structure:**

```yaml
current_profile: default
profiles:
  default:
    server: cubeapm.example.com
    query_port: 3140
    ingest_port: 3130
    admin_port: 3199
    token: your-api-token
    output: table
    read_only: false
```

**Field reference:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `current_profile` | string | `""` | Name of the active profile |
| `profiles.<name>.server` | string | `""` | CubeAPM server hostname or IP |
| `profiles.<name>.query_port` | int | `3140` | Port for trace/metric/log queries |
| `profiles.<name>.ingest_port` | int | `3130` | Port for data ingestion |
| `profiles.<name>.admin_port` | int | `3199` | Port for admin operations |
| `profiles.<name>.token` | string | `""` | API authentication token |
| `profiles.<name>.output` | string | `table` | Default output format: `table`, `json`, `yaml` |
| `profiles.<name>.read_only` | bool | `false` | Block write/delete operations (safety mode) |

**File permissions:** The config file is created with `0600` permissions (owner read/write only) and the config directory with `0700` permissions. This protects tokens from being readable by other users on the system.

**View the current config:**

```bash
cubeapm config view
```

**Set individual values:**

```bash
cubeapm config set server cubeapm.example.com
cubeapm config set token my-api-token
cubeapm config set output json
cubeapm config set query_port 3140
cubeapm config set ingest_port 3130
cubeapm config set admin_port 3199
```

**Read individual values:**

```bash
cubeapm config get server
cubeapm config get token
cubeapm config get current_profile
```

### Environment Variables

All settings can be provided or overridden via environment variables. These take precedence over the config file but are overridden by CLI flags.

| Environment Variable | Type | Default | Description |
|---------------------|------|---------|-------------|
| `CUBEAPM_SERVER` | string | | CubeAPM server hostname or IP |
| `CUBEAPM_TOKEN` | string | | API authentication token |
| `CUBEAPM_QUERY_PORT` | int | `3140` | Query API port |
| `CUBEAPM_INGEST_PORT` | int | `3130` | Ingest API port |
| `CUBEAPM_ADMIN_PORT` | int | `3199` | Admin API port |
| `CUBEAPM_READ_ONLY` | bool | `false` | Block write/delete operations |
| `CUBEAPM_NO_INPUT` | bool | `false` | Disable interactive prompts |
| `CUBEAPM_QUIET` | bool | `false` | Suppress informational output |

**Example: setting up a shell profile for persistent access:**

```bash
# Add to ~/.bashrc or ~/.zshrc
export CUBEAPM_SERVER=cubeapm.example.com
export CUBEAPM_TOKEN=your-api-token
```

**Example: CI pipeline snippet (GitHub Actions):**

```yaml
env:
  CUBEAPM_SERVER: ${{ secrets.CUBEAPM_SERVER }}
  CUBEAPM_TOKEN: ${{ secrets.CUBEAPM_TOKEN }}

steps:
  - name: Check service health
    run: cubeapm traces services -o json
```

**Example: CI pipeline snippet (GitLab CI):**

```yaml
variables:
  CUBEAPM_SERVER: $CUBEAPM_SERVER
  CUBEAPM_TOKEN: $CUBEAPM_TOKEN

check-services:
  script:
    - cubeapm traces services -o json
```

### Multiple Profiles

You can configure multiple profiles for different CubeAPM instances (e.g., development, staging, production) and switch between them.

**Create profiles with interactive login:**

```bash
# Create a "dev" profile
cubeapm login
# Profile name [default]: dev
# CubeAPM server: localhost
# API token (leave empty for no auth):
# ...

# Create a "staging" profile
cubeapm login
# Profile name [default]: staging
# CubeAPM server: cubeapm.staging.example.com
# API token (leave empty for no auth): staging-token
# ...

# Create a "production" profile
cubeapm login
# Profile name [default]: production
# CubeAPM server: cubeapm.prod.example.com
# API token (leave empty for no auth): prod-token
# ...
```

**List all profiles:**

```bash
cubeapm config profiles list
```

Output shows the active profile marked with `*`:

```
  dev
  staging
* production
```

**Switch the active profile:**

```bash
cubeapm config profiles use staging
```

**Use a specific profile for a single command (without switching):**

```bash
cubeapm --profile production traces services
cubeapm --profile dev logs query 'error' --last 1h
```

**Delete a profile:**

```bash
cubeapm config profiles delete staging
```

**Resulting config file with multiple profiles:**

```yaml
current_profile: production
profiles:
  dev:
    server: localhost
    output: table
  staging:
    server: cubeapm.staging.example.com
    token: staging-token
    output: json
  production:
    server: cubeapm.prod.example.com
    query_port: 3140
    ingest_port: 3130
    admin_port: 3199
    token: prod-token
    output: table
    read_only: true
```

---

## Deployment Scenarios

### Self-Hosted (Single Binary)

When CubeAPM is installed directly on a server as a single binary, the CLI connects directly to the server's address and ports.

**Setup:**

```bash
cubeapm login
# CubeAPM server: cubeapm.internal.example.com
# API token (leave empty for no auth): your-token
# Query port [3140]: 3140
# Ingest port [3130]: 3130
# Admin port [3199]: 3199
```

**Port configuration considerations:**

If CubeAPM is configured with non-default ports, you need to specify them during login or via config:

```bash
cubeapm config set query_port 8080
cubeapm config set ingest_port 8081
cubeapm config set admin_port 8082
```

**Firewall considerations:**

Ensure the following ports are open between the CLI host and the CubeAPM server:

```bash
# Minimum for read-only access
# TCP: server:3140 (query)

# For data ingestion
# TCP: server:3130 (ingest)

# For admin operations (log deletion)
# TCP: server:3199 (admin)
```

### Docker Deployment

When CubeAPM runs in Docker, you need to account for port mapping between the container and the host.

**Standard docker-compose setup:**

```yaml
# docker-compose.yml
services:
  cubeapm:
    image: cubeapm/cubeapm:latest
    ports:
      - "3125:3125"   # Web UI
      - "3130:3130"   # Ingest
      - "3140:3140"   # Query API
      - "3199:3199"   # Admin
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
```

**CLI configuration (from the Docker host):**

```bash
cubeapm login
# CubeAPM server: localhost
# Query port [3140]: 3140
# Ingest port [3130]: 3130
# Admin port [3199]: 3199
```

**Custom port mapping:**

If you map CubeAPM ports to different host ports:

```yaml
# docker-compose.yml
services:
  cubeapm:
    image: cubeapm/cubeapm:latest
    ports:
      - "9140:3140"   # Query API mapped to host port 9140
      - "9130:3130"   # Ingest mapped to host port 9130
      - "9199:3199"   # Admin mapped to host port 9199
```

Configure the CLI to use the host-side ports:

```bash
cubeapm login
# CubeAPM server: localhost
# Query port [3140]: 9140
# Ingest port [3130]: 9130
# Admin port [3199]: 9199
```

**Accessing from a remote machine:**

If CubeAPM runs in Docker on a remote server, use that server's hostname:

```bash
cubeapm login
# CubeAPM server: docker-host.example.com
# Query port [3140]: 3140
```

**Docker network considerations:**

If the CLI runs inside another container on the same Docker network, use the service name:

```bash
export CUBEAPM_SERVER=cubeapm
export CUBEAPM_QUERY_PORT=3140
```

### Kubernetes Deployment

When CubeAPM runs in Kubernetes, there are several ways to connect the CLI.

**Option A: Port-forward for local CLI access (recommended for ad-hoc use)**

Forward the three ports the CLI needs:

```bash
# Forward query port
kubectl port-forward -n monitoring svc/cubeapm 3140:3140 &

# Forward ingest port (if you need to push data)
kubectl port-forward -n monitoring svc/cubeapm 3130:3130 &

# Forward admin port (if you need log deletion)
kubectl port-forward -n monitoring svc/cubeapm 3199:3199 &

# Configure CLI to use localhost
cubeapm login
# CubeAPM server: localhost
```

To forward all three ports in a single command:

```bash
kubectl port-forward -n monitoring svc/cubeapm 3140:3140 3130:3130 3199:3199
```

**Option B: In-cluster access (from a pod)**

If the CLI runs inside a pod in the same cluster, use the Kubernetes Service DNS name:

```bash
export CUBEAPM_SERVER=cubeapm.monitoring.svc.cluster.local
cubeapm traces services
```

**Option C: Via Ingress or LoadBalancer**

If CubeAPM is exposed via an Ingress or LoadBalancer, configure the CLI to use the external address. Note that the Ingress/LoadBalancer may route all ports through a single address, or expose them on different ports (see the Load Balancer section below).

```bash
cubeapm login
# CubeAPM server: cubeapm.example.com
# Query port [3140]: 443    # if behind HTTPS ingress
```

**Useful kubectl commands for discovering CubeAPM:**

```bash
# Find CubeAPM pods
kubectl get pods -n monitoring -l app=cubeapm

# Find CubeAPM services and their ports
kubectl get svc -n monitoring -l app=cubeapm

# Check CubeAPM pod logs for port configuration
kubectl logs -n monitoring deployment/cubeapm | head -20

# Describe the service to see port mappings
kubectl describe svc -n monitoring cubeapm
```

---

## Edge Cases & Troubleshooting

### Custom Ports

If your CubeAPM instance uses non-default ports, you must configure them in the CLI.

**How to discover current ports:**

Check the CubeAPM server startup logs or configuration:

```bash
# If running as a binary, check the process
ps aux | grep cubeapm
# or check the config file used to start CubeAPM

# If running in Docker
docker logs cubeapm 2>&1 | head -20

# If running in Kubernetes
kubectl logs -n monitoring deployment/cubeapm 2>&1 | head -20
```

**How to configure custom ports in the CLI:**

```bash
# During login
cubeapm login
# Query port [3140]: 8080
# Ingest port [3130]: 8081
# Admin port [3199]: 8082

# Via config set
cubeapm config set query_port 8080
cubeapm config set ingest_port 8081
cubeapm config set admin_port 8082

# Via environment variables
export CUBEAPM_QUERY_PORT=8080
export CUBEAPM_INGEST_PORT=8081
export CUBEAPM_ADMIN_PORT=8082

# Via CLI flags (per-command)
cubeapm --query-port 8080 traces services
cubeapm --ingest-port 8081 ingest metrics --file metrics.txt
cubeapm --admin-port 8082 logs delete list
```

### CubeAPM Behind a Load Balancer

When CubeAPM sits behind a load balancer or reverse proxy, port mapping may differ from the defaults.

**Single entry point with port routing:**

If the load balancer maps different external ports to CubeAPM's internal ports:

```bash
# Load balancer at lb.example.com
# External port 443 → CubeAPM 3140 (query)
# External port 8443 → CubeAPM 3130 (ingest)
# External port 9443 → CubeAPM 3199 (admin)

cubeapm login
# CubeAPM server: lb.example.com
# Query port [3140]: 443
# Ingest port [3130]: 8443
# Admin port [3199]: 9443
```

**Health check configuration:**

If your load balancer needs a health check endpoint, CubeAPM's query port serves the API and can be used for health checks:

```bash
# Health check URL for the query API
curl -s http://cubeapm.example.com:3140/api/services | head -1
```

**Multiple CubeAPM instances behind a load balancer:**

The CLI treats the load balancer address as the server. All requests go through the load balancer, which distributes them to backend CubeAPM instances:

```bash
export CUBEAPM_SERVER=cubeapm-lb.example.com
cubeapm traces services
```

### Network Issues

**Testing connectivity to each port independently:**

Use `curl` to verify that each port is reachable before troubleshooting the CLI:

```bash
# Test query port (should return JSON with services list)
curl -s -o /dev/null -w "%{http_code}" http://cubeapm.example.com:3140/api/services
# Expected: 200

# Test ingest port
curl -s -o /dev/null -w "%{http_code}" http://cubeapm.example.com:3130/
# Expected: 200 or 404 (server is reachable)

# Test admin port
curl -s -o /dev/null -w "%{http_code}" http://cubeapm.example.com:3199/
# Expected: 200 or 404 (server is reachable)

# Test with authentication token
curl -s -H "Authorization: Bearer your-token" http://cubeapm.example.com:3140/api/services
```

**Firewall blocking ports:**

If `curl` times out or refuses connection, check firewall rules:

```bash
# Check if the port is open (macOS/Linux)
nc -zv cubeapm.example.com 3140

# Check with timeout
timeout 5 bash -c 'echo > /dev/tcp/cubeapm.example.com/3140' && echo "Open" || echo "Closed"
```

**DNS resolution issues:**

```bash
# Verify DNS resolves
nslookup cubeapm.example.com
# or
dig cubeapm.example.com
```

**Using verbose mode for debugging:**

The CLI's `--verbose` flag logs every HTTP request and response, which is invaluable for diagnosing connection issues:

```bash
cubeapm --verbose traces services
```

Output:

```
> GET http://cubeapm.example.com:3140/api/services
> Authorization: [REDACTED]
< 200 200 OK
< Content-Type: application/json
```

### Common Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `server address not configured; run 'cubeapm login' or set CUBEAPM_SERVER` | No server configured in any source | Run `cubeapm login` or `export CUBEAPM_SERVER=...` |
| `authentication failed (HTTP 401): check your token with 'cubeapm login'` | Invalid, expired, or missing token when server requires auth | Verify token with `cubeapm config get token`; re-run `cubeapm login` |
| `access denied (HTTP 403): your token may not have sufficient permissions` | Token is valid but lacks permission for the operation | Use a token with broader permissions |
| `resource not found (HTTP 404)` | Wrong port, wrong endpoint, or CubeAPM version mismatch | Verify you are connecting to the correct port for the operation |
| `request failed: dial tcp ...: connect: connection refused` | CubeAPM is not running or wrong server/port | Verify server address and that CubeAPM is running; test with `curl` |
| `request failed: dial tcp ...: i/o timeout` | Firewall blocking the port or wrong address | Check firewall rules; test connectivity with `nc -zv host port` |
| `request failed: dial tcp ...: no such host` | DNS cannot resolve the server hostname | Verify the hostname; try using an IP address instead |
| `login requires interactive prompts; cannot run with --no-input` | Running `cubeapm login` in non-interactive mode | Use `cubeapm config set` or environment variables instead of `login` |
| `command '...' is blocked in read-only mode` | `read_only: true` is set in the profile or via `--read-only` | Use `--read-only=false` or remove `read_only` from the profile |
| `invalid server address "..."` | Malformed server address (e.g., contains invalid characters) | Provide a valid hostname or IP address |
| `profile "..." does not exist` | Trying to use or switch to a profile that has not been created | Run `cubeapm config profiles list` to see available profiles; create one with `cubeapm login` |
| `no active profile set` | No current profile in config and no env vars/flags set | Run `cubeapm login` or `cubeapm config profiles use <name>` |

### Timeout and Connection Tuning

The CLI uses these default timeouts (not user-configurable):

| Timeout | Value | Description |
|---------|-------|-------------|
| Connection dial | 10 seconds | Time to establish TCP connection |
| TLS handshake | 10 seconds | Time for TLS negotiation (HTTPS) |
| Request total | 60 seconds | Maximum time for an entire HTTP request |
| Keep-alive probe | 30 seconds | TCP keep-alive interval |

If you experience timeouts, the issue is likely network-related (firewall, routing) rather than a CLI configuration problem.

---

## Security Best Practices

### Protect Your Tokens

**Store tokens in environment variables, not in scripts or command history:**

```bash
# Good: environment variable (set in .bashrc/.zshrc or CI secrets)
export CUBEAPM_TOKEN=your-token

# Good: the config file is created with 0600 permissions
cubeapm login

# Bad: token visible in command history
cubeapm --token my-secret-token traces services  # avoid in production

# Bad: token hardcoded in a script
TOKEN="my-secret-token"  # avoid this
```

**In CI/CD, use your platform's secrets management:**

```yaml
# GitHub Actions
env:
  CUBEAPM_TOKEN: ${{ secrets.CUBEAPM_TOKEN }}

# GitLab CI - use protected/masked variables
variables:
  CUBEAPM_TOKEN: $CUBEAPM_TOKEN_SECRET
```

### Use Read-Only Mode for Safety

Enable `read_only` mode to prevent accidental data modification or deletion. This is especially recommended for:

- Automated scripts that should only query data
- Profiles used by coding agents (LLMs)
- Shared team profiles

```bash
# Set read-only in the profile
cubeapm config set read_only true

# Or via environment variable
export CUBEAPM_READ_ONLY=true

# Or via CLI flag for a single command
cubeapm --read-only traces services
```

When read-only mode is active, write operations like `ingest metrics`, `ingest logs`, and `logs delete run` are blocked with an error message.

### Restrict Admin Port Access

The admin port (default 3199) provides powerful operations like log deletion. Limit access to this port:

- Only open the admin port to trusted networks or specific IP addresses
- Use a separate profile with admin access and keep it distinct from day-to-day profiles
- Consider not configuring the admin port in profiles used by automated systems

```yaml
# Day-to-day profile (no admin access needed)
profiles:
  daily:
    server: cubeapm.example.com
    token: read-token
    read_only: true

  # Admin profile (restricted use)
  admin:
    server: cubeapm.example.com
    token: admin-token
    admin_port: 3199
```

### Network Segmentation for Ports

Consider restricting network access to each CubeAPM port based on the principle of least privilege:

| Port | Who Needs Access |
|------|-----------------|
| 3125 (Web UI) | Developers, SREs (browser) |
| 3130 (Ingest) | Application servers, agents, CI pipelines |
| 3140 (Query) | Developers, SREs, dashboards, CLI users |
| 3199 (Admin) | SRE team leads, on-call engineers only |
| 4317/4318 (OTLP) | Application servers running OpenTelemetry SDKs |

### Verify Config File Permissions

The CLI creates the config file with secure permissions, but verify periodically:

```bash
# Check permissions (should be -rw-------)
ls -la ~/.config/cubeapm-cli/config.yaml

# Fix if needed
chmod 600 ~/.config/cubeapm-cli/config.yaml
chmod 700 ~/.config/cubeapm-cli/
```

### Audit Your Profiles

Periodically review saved profiles and remove unused ones:

```bash
# List all profiles
cubeapm config profiles list

# View the full config including tokens
cubeapm config view

# Remove unused profiles
cubeapm config profiles delete old-staging
```

---

## Complete CLI Flags Reference

These global flags are available on every command and relate to authentication and connection:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--server` | | string | | CubeAPM server address (hostname or IP) |
| `--token` | | string | | Authentication token |
| `--query-port` | | int | `3140` | Query API port |
| `--ingest-port` | | int | `3130` | Ingest API port |
| `--admin-port` | | int | `3199` | Admin API port |
| `--profile` | | string | | Use a specific connection profile |
| `--read-only` | | bool | `false` | Block write/delete operations |
| `--no-input` | | bool | `false` | Disable interactive prompts (CI/agent use) |
| `--quiet` | `-q` | bool | `false` | Suppress informational output |
| `--verbose` | | bool | `false` | Log HTTP requests and responses |
| `--output` | `-o` | string | `table` | Output format: `table`, `json`, `yaml` |
| `--no-color` | | bool | `false` | Disable colored output |
