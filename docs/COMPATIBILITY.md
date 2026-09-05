# CubeAPM compatibility and maintenance

Checked 2026-09-06. These are documentation and local-test baselines, not certification against every server deployment.

## Build requirements

- Go 1.26 or later; this checkout selects Go 1.27.1 using `toolchain` in `cli-go/go.mod`.
- Go 1.27 builds for macOS require macOS 13 or later. See the [Go 1.27 release notes](https://go.dev/doc/go1.27).
- Docs: Node.js 24 or later and pnpm 11.25.0. Install with `pnpm install --frozen-lockfile` in `web/`.
- YAML uses the maintained [YAML organization v3 implementation](https://github.com/yaml/go-yaml), preserving the v3 configuration API.

## Upstream API baseline

The native trace API uses `/api/traces/api/v1/search` with `query`, `env`, `service`, and Unix-second `start`/`end` parameters. Trace fetch uses `/api/traces/api/v1/traces/<id>` with the same time units. Responses are normalized for the CLI's existing trace output.

Reference: [CubeAPM trace API](https://docs.cubeapm.com/http-apis/traces-apis), checked 2026-09-06. CubeAPM publishes rolling documentation rather than a release-specific API manual. Optional Jaeger-derived endpoints and extra filters remain deployment-dependent; native request parameters are the default.

## Maintaining this baseline

Dependabot is configured to propose weekly Go, npm, and GitHub Actions updates. Review migrations before merging; CI verifies the Go tests, race checks, security/static analysis, command documentation, and docs build.

Run `go list -m -u all` from `cli-go/` and `pnpm outdated` from `web/` to check newer releases. Commit `go.mod` with `go.sum`, and `package.json` with `pnpm-lock.yaml`. Keep `fumadocs-core` and the `fumadocs-ui` alias to `@fumadocs/base-ui` on matching versions.

The website's static search uses Fumadocs' default search engine. After building, run `pnpm test:search` to verify that the exported index can be loaded and queried by the installed client.
