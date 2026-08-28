# cubeapm MCP server

Exposes the `cubeapm` CLI to any [MCP](https://modelcontextprotocol.io)-compatible client — most notably **Claude Desktop**, which (unlike Claude Code) has no shell access and can't run the CLI directly.

This server shells out to your locally-installed, already-authenticated `cubeapm` CLI — it doesn't reimplement any of the CubeAPM API client. Every call is forced to run with `--read-only --no-input`, so it can only query data, never mutate CubeAPM state.

## Requirements

- Node.js 18+
- `cubeapm` installed and on your `PATH` (see the [root README](../README.md#installation)), authenticated via `cubeapm login`

## Install

```bash
cd mcp-server
npm install
```

## Configure in Claude Desktop

Add to your `claude_desktop_config.json` (macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "cubeapm": {
      "command": "node",
      "args": ["/absolute/path/to/cubeapm-cli/mcp-server/index.js"]
    }
  }
}
```

Restart Claude Desktop. If `cubeapm` isn't on the `PATH` Claude Desktop launches with, set an explicit path:

```json
{
  "mcpServers": {
    "cubeapm": {
      "command": "node",
      "args": ["/absolute/path/to/cubeapm-cli/mcp-server/index.js"],
      "env": { "CUBEAPM_BIN": "/full/path/to/cubeapm" }
    }
  }
}
```

## What it exposes

**Tools** (callable automatically by the model when relevant):

- `cubeapm_traces_search` — search traces by service, status, duration, tags, time range
- `cubeapm_traces_get` — fetch a trace's full waterfall by ID
- `cubeapm_traces_services` — list services reporting telemetry
- `cubeapm_logs_query` — LogsQL query
- `cubeapm_metrics_query` — instant PromQL query

**Prompt** (shows up as a slash command in MCP clients that support prompts, e.g. `/cubeapm:usecube` in Claude Desktop):

- `usecube` — takes one optional `request` argument describing what you want to investigate; routes to whichever tool(s) above fit.

## Notes

- This is a thin process-spawning wrapper (`node:child_process.execFile`, no shell interpolation) — it never passes user input through a shell.
- Not related to the [`cubeapm/` skill](../cubeapm/) shipped for coding agents with shell access (Claude Code, Cursor, etc.) — use that one directly if your agent can already run CLI commands. This MCP server is specifically for clients that can't.
