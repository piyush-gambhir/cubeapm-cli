#!/usr/bin/env node
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { execFile } from "node:child_process";

// Resolved via PATH by default (execFile uses execvp-style lookup for
// bare command names). Override with CUBEAPM_BIN if it's not on PATH.
const CUBEAPM_BIN = process.env.CUBEAPM_BIN || "cubeapm";

function runCubeapm(args) {
  const fullArgs = [...args, "-o", "json", "--no-input", "--read-only"];
  return new Promise((resolve) => {
    execFile(
      CUBEAPM_BIN,
      fullArgs,
      { timeout: 30_000, maxBuffer: 10 * 1024 * 1024 },
      (error, stdout, stderr) => {
        if (error) {
          const hint =
            error.code === "ENOENT"
              ? "\n\ncubeapm binary not found on PATH. Install it (see https://github.com/piyush-gambhir/cubeapm-cli#installation) or set CUBEAPM_BIN to its full path."
              : "";
          resolve({
            isError: true,
            content: [
              {
                type: "text",
                text: `cubeapm command failed: ${error.message}\n${stderr || ""}`.trim() + hint,
              },
            ],
          });
          return;
        }
        resolve({
          content: [{ type: "text", text: stdout || "(no output)" }],
        });
      },
    );
  });
}

function addTimeFlags(args, { last, from, to }) {
  if (last) args.push("--last", last);
  if (from) args.push("--from", from);
  if (to) args.push("--to", to);
}

const server = new McpServer({
  name: "cubeapm",
  version: "1.0.0",
});

server.registerTool(
  "cubeapm_traces_search",
  {
    title: "Search CubeAPM traces",
    description:
      "Search distributed traces in CubeAPM (Jaeger-compatible) by service, status, duration, tags, etc.",
    inputSchema: {
      service: z.string().describe('Service name, e.g. "api-gateway"'),
      env: z.string().optional().describe('Environment, e.g. "PROD" or "UAT" (upper-case)'),
      query: z.string().optional().describe('Operation name filter, e.g. "GET /api/users"'),
      status: z.enum(["error", "ok"]).optional(),
      minDuration: z.string().optional().describe('e.g. "500ms", "1s"'),
      maxDuration: z.string().optional().describe('e.g. "5s"'),
      tags: z.array(z.string()).optional().describe('Span tag filters as "key=value" pairs'),
      spanKind: z.enum(["client", "server", "producer", "consumer", "internal"]).optional(),
      limit: z.number().int().positive().optional(),
      last: z.string().optional().describe('Relative time window, e.g. "1h", "30m", "2d"'),
      from: z.string().optional(),
      to: z.string().optional(),
    },
  },
  async ({ service, env, query, status, minDuration, maxDuration, tags, spanKind, limit, last, from, to }) => {
    const args = ["traces", "search", "--service", service];
    if (env) args.push("--env", env);
    if (query) args.push("--query", query);
    if (status) args.push("--status", status);
    if (minDuration) args.push("--min-duration", minDuration);
    if (maxDuration) args.push("--max-duration", maxDuration);
    if (spanKind) args.push("--span-kind", spanKind);
    if (limit) args.push("--limit", String(limit));
    for (const tag of tags ?? []) args.push("--tags", tag);
    addTimeFlags(args, { last, from, to });
    return runCubeapm(args);
  },
);

server.registerTool(
  "cubeapm_traces_get",
  {
    title: "Get a CubeAPM trace by ID",
    description: "Retrieve the full waterfall/span data for a specific trace ID.",
    inputSchema: {
      traceId: z.string().describe("Hex trace ID, typically obtained from cubeapm_traces_search"),
      last: z.string().optional(),
      from: z.string().optional(),
      to: z.string().optional(),
    },
  },
  async ({ traceId, last, from, to }) => {
    const args = ["traces", "get", traceId];
    addTimeFlags(args, { last, from, to });
    return runCubeapm(args);
  },
);

server.registerTool(
  "cubeapm_traces_services",
  {
    title: "List services reporting to CubeAPM",
    description: "List all services that have reported traces/telemetry to CubeAPM.",
    inputSchema: {
      env: z.string().optional().describe('e.g. "PROD" or "UAT"'),
      last: z.string().optional(),
      from: z.string().optional(),
      to: z.string().optional(),
    },
  },
  async ({ env, last, from, to }) => {
    const args = ["traces", "services"];
    if (env) args.push("--env", env);
    addTimeFlags(args, { last, from, to });
    return runCubeapm(args);
  },
);

server.registerTool(
  "cubeapm_logs_query",
  {
    title: "Query CubeAPM logs",
    description:
      "Query logs using LogsQL syntax (VictoriaLogs-compatible). Supports keyword search, field filters, AND/OR/NOT, regex, etc.",
    inputSchema: {
      logsql: z.string().describe('LogsQL expression, e.g. "error AND service:api"'),
      service: z.string().optional(),
      level: z.enum(["error", "warn", "info", "debug"]).optional(),
      stream: z.string().optional().describe('e.g. \'{host="web-1"}\''),
      limit: z.number().int().positive().optional(),
      last: z.string().optional(),
      from: z.string().optional(),
      to: z.string().optional(),
    },
  },
  async ({ logsql, service, level, stream, limit, last, from, to }) => {
    const args = ["logs", "query", logsql];
    if (service) args.push("--service", service);
    if (level) args.push("--level", level);
    if (stream) args.push("--stream", stream);
    if (limit) args.push("--limit", String(limit));
    addTimeFlags(args, { last, from, to });
    return runCubeapm(args);
  },
);

server.registerTool(
  "cubeapm_metrics_query",
  {
    title: "Query CubeAPM metrics (PromQL)",
    description: "Execute an instant PromQL query against CubeAPM's Prometheus-compatible metrics API.",
    inputSchema: {
      promql: z.string().describe('PromQL expression, e.g. "up" or "rate(http_requests_total[5m])"'),
      time: z.string().optional().describe('Evaluation timestamp, e.g. "now-1h"'),
    },
  },
  async ({ promql, time }) => {
    const args = ["metrics", "query", promql];
    if (time) args.push("--time", time);
    return runCubeapm(args);
  },
);

server.registerPrompt(
  "usecube",
  {
    title: "Use CubeAPM",
    description: "Activate CubeAPM observability tools (traces/logs/metrics) for this request",
    argsSchema: {
      request: z
        .string()
        .optional()
        .describe('What you want to investigate, e.g. "errors in dealer-leads-backend last hour"'),
    },
  },
  ({ request }) => ({
    messages: [
      {
        role: "user",
        content: {
          type: "text",
          text: request
            ? `Using the cubeapm tools (traces search/get, list services, logs query, metrics query), help with: ${request}`
            : `Use the cubeapm tools (traces search/get, list services, logs query, metrics query) to help answer what I ask next.`,
        },
      },
    ],
  }),
);

const transport = new StdioServerTransport();
await server.connect(transport);
