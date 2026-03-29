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
CubeAPM server: cube.example.com

Authentication:
  1. Email/Password
  2. No authentication
Choose [1]: 1
Email: user@example.com
Password:

Authenticating... OK (session expires 2024-01-16 18:00 UTC)
Connection verified (12 services found)

Query port [3140]:
Ingest port [3130]:
Admin port [3199]:

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
export CUBEAPM_SERVER=cube.example.com
export CUBEAPM_EMAIL=user@example.com
export CUBEAPM_PASSWORD=your-password

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
cubeapm --server cube.example.com --email user@example.com --password secret traces services
```

---

## Getting Your Credentials

### Step 1: Find Your CubeAPM Server Address

The server address is the hostname or URL where your CubeAPM instance is running.

**Self-hosted (single binary):**

If you installed CubeAPM directly on a server, use that server's hostname or IP:

```bash
cubeapm login
# CubeAPM server: cubeapm.internal.example.com
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

Use the Kubernetes service name (if connecting from within the cluster) or the external ingress/load balancer address:

```bash
# In-cluster
# CubeAPM server: cubeapm.monitoring.svc.cluster.local

# External
# CubeAPM server: cube.example.com
```

**Reverse proxy / HTTPS:**

If CubeAPM is behind a reverse proxy (like nginx) serving on HTTPS:

```bash
# CubeAPM server: https://cube.example.com
```

When using a reverse proxy, all ports (query, ingest, admin) are typically served through the same host on port 443. In this case, use the same port for all three when prompted.

### Step 2: Find Your Port Configuration

CubeAPM uses three ports for different purposes:

| Port | Default | Purpose |
|------|---------|---------|
| Query | 3140 | Traces, metrics, and logs queries |
| Ingest | 3130 | Pushing metrics and logs data |
| Admin | 3199 | Administrative operations (log deletion) |

**Standard deployment (direct access):** Use the defaults (3140, 3130, 3199).

**Reverse proxy deployment:** If CubeAPM is behind a reverse proxy on HTTPS, you typically don't need custom ports -- the reverse proxy routes everything through port 443. Set all three ports to 443 or leave as defaults if the proxy handles the routing.

### Step 3: Get Your Login Credentials

CubeAPM uses email and password authentication (powered by Ory Kratos). You need the email and password of a user account on your CubeAPM instance.

**If authentication is enabled (most production deployments):**

1. Open your CubeAPM web UI in a browser
2. Log in with your email and password
3. Use the same email and password for the CLI

**If authentication is disabled (local/dev instances):**

Some CubeAPM instances (especially local development setups) run without authentication. Choose "No authentication" during `cubeapm login`.

**Getting an account:**

If you don't have an account:
- Ask your CubeAPM administrator to create one
- Or self-register at the CubeAPM web UI (if self-service signup is enabled)

---

## Authentication Methods

### Email/Password Authentication

This is the primary authentication method for CubeAPM instances with authentication enabled. The CLI performs the full login flow:

1. Initiates a login session with the CubeAPM server
2. Submits your email and password
3. Receives a session cookie (valid for 24 hours by default)
4. Caches the session cookie for subsequent commands
5. Automatically re-authenticates when the session expires

```bash
# Interactive
cubeapm login
# Choose: 1. Email/Password
# Enter email and password

# Non-interactive (via environment variables)
export CUBEAPM_SERVER=cube.example.com
export CUBEAPM_EMAIL=user@example.com
export CUBEAPM_PASSWORD=your-password
cubeapm traces services
```

**Session management:**

- Sessions are cached in the config file and reused across CLI invocations
- When a session expires, the CLI automatically re-authenticates using stored credentials
- You can force re-authentication by running `cubeapm login` again

### No Authentication

For CubeAPM instances running without authentication (common in development or internal networks with network-level access control):

```bash
cubeapm login
# Choose: 2. No authentication
```

No credentials are needed. The CLI connects directly to the CubeAPM API ports.

---

## Configuration Priority

Settings are resolved in this order (highest priority first):

1. **CLI flags** (`--server`, `--email`, `--password`, `--query-port`, etc.)
2. **Environment variables** (`CUBEAPM_SERVER`, `CUBEAPM_EMAIL`, `CUBEAPM_PASSWORD`, etc.)
3. **Profile configuration** (`~/.config/cubeapm-cli/config.yaml`)

### Configuration File

Location: `~/.config/cubeapm-cli/config.yaml` (or `$XDG_CONFIG_HOME/cubeapm-cli/config.yaml`)

Example:

```yaml
current_profile: production
profiles:
  production:
    server: cube.example.com
    auth_method: kratos
    email: admin@example.com
    password: your-password
    session_cookie: "ory_kratos_session=..."
    session_expiry: "2024-01-16T18:00:00Z"
    query_port: 3140
    ingest_port: 3130
    admin_port: 3199
  local:
    server: localhost
    auth_method: none
    query_port: 3140
    ingest_port: 3130
    admin_port: 3199
```

The config file is created with `0600` permissions (owner read/write only) to protect stored credentials.

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CUBEAPM_SERVER` | CubeAPM server address |
| `CUBEAPM_EMAIL` | Login email |
| `CUBEAPM_PASSWORD` | Login password |
| `CUBEAPM_QUERY_PORT` | Query API port (default: 3140) |
| `CUBEAPM_INGEST_PORT` | Ingest API port (default: 3130) |
| `CUBEAPM_ADMIN_PORT` | Admin API port (default: 3199) |
| `CUBEAPM_READ_ONLY` | Block write/delete operations (`true`/`false`) |
| `CUBEAPM_NO_INPUT` | Disable interactive prompts (`true`/`false`) |
| `CUBEAPM_QUIET` | Suppress informational output (`true`/`false`) |

### CLI Flags

| Flag | Description |
|------|-------------|
| `--server <addr>` | CubeAPM server address |
| `--email <email>` | Login email |
| `--password <password>` | Login password |
| `--query-port <port>` | Query port (default: 3140) |
| `--ingest-port <port>` | Ingest port (default: 3130) |
| `--admin-port <port>` | Admin port (default: 3199) |
| `--profile <name>` | Use a specific connection profile |
| `--verbose` | Enable verbose HTTP request logging |
| `--read-only` | Block write/delete operations |

---

## Multiple Profiles

Profiles let you manage connections to multiple CubeAPM instances:

```bash
# Create profiles via login
cubeapm login                                 # saves as "default"
cubeapm login --server cube-staging.example.com  # saves as another profile

# List all profiles
cubeapm config profiles list

# Switch active profile
cubeapm config profiles use staging

# Use a profile for a single command
cubeapm traces services --profile production

# Delete a profile
cubeapm config profiles delete old-profile
```

---

## CI/CD & Automation

### GitHub Actions

```yaml
- name: Check service health
  env:
    CUBEAPM_SERVER: cube.example.com
    CUBEAPM_EMAIL: ${{ secrets.CUBEAPM_EMAIL }}
    CUBEAPM_PASSWORD: ${{ secrets.CUBEAPM_PASSWORD }}
  run: |
    cubeapm traces services -o json
    cubeapm metrics query 'up' -o json
```

### Non-Interactive Mode

Use `--no-input` (or `CUBEAPM_NO_INPUT=true`) to disable all interactive prompts. The CLI will fail with an error if credentials are missing rather than prompting:

```bash
export CUBEAPM_NO_INPUT=true
export CUBEAPM_SERVER=cube.example.com
export CUBEAPM_EMAIL=ci-user@example.com
export CUBEAPM_PASSWORD=ci-password
cubeapm traces services -o json
```

---

## Troubleshooting

### Common Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `authentication failed (HTTP 401)` | Invalid email/password or expired session | Run `cubeapm login` to re-authenticate |
| `connection refused` | Wrong server address or port | Verify server address and ports with `cubeapm config view` |
| `connection timeout` | Server unreachable or firewall blocking | Check network connectivity and firewall rules |
| `server address not configured` | No server set | Run `cubeapm login` or set `CUBEAPM_SERVER` |
| `login failed: invalid email or password` | Wrong credentials | Double-check email and password at the CubeAPM web UI |

### Debugging

Use `--verbose` to see the full HTTP request/response cycle (auth headers are redacted):

```bash
cubeapm traces services --verbose
```

Check your current configuration:

```bash
cubeapm config view
cubeapm config get server
cubeapm config get email
cubeapm config get auth_method
```

### Testing Connectivity

```bash
# Verify the server is reachable
curl -s -o /dev/null -w "%{http_code}" https://cube.example.com/

# Verify the query API port
curl -s https://cube.example.com:3140/api/traces/api/v1/services
```

---

## Security Best Practices

1. **Config file permissions:** The CLI creates the config file with `0600` permissions. Do not loosen these.
2. **Environment variables in CI:** Store credentials as CI secrets, never in plaintext in pipeline configs.
3. **Read-only mode:** Use `--read-only` or `CUBEAPM_READ_ONLY=true` for agent/automation contexts to prevent accidental data modifications.
4. **Password storage:** The password is stored in the config file for automatic session renewal. For higher-security environments, use environment variables instead and don't store the password in the config.
5. **Avoid CLI flag credentials in shell history:** Prefer `cubeapm login` or environment variables over `--email`/`--password` flags to avoid credentials appearing in shell history.
