# CubeAPM Node tooling review

Reviewed locally on 2026-10-26, based on `origin/main` at `ef9ce3a`.
Upstream security configuration, action pins, bounded dependency overrides and
release-age exceptions are preserved byte for byte where unchanged by Node pins.

- Docs retain pnpm and their lockfile, use Node 26.12.1/pnpm 12.15.0, matching
  Fumadocs core/UI 16.17.8 and MDX 15.4.10. TypeScript 7.0.2 works standalone.
- MCP retains npm, its separate lockfile and version 1.0.0. Its SDK updates from
  1.32.1 to 1.32.2. Both packages retain one existing `.nvmrc` each, now 26.12.1.
- Docs frozen install, peers, types, 48-page static build and installation search
  pass. Wrangler dry run reads 274 static assets without uploading anything.
- Local workerd browser validation exercises the installation route, hydrated
  search dialog, real static search index and result navigation.
- MCP clean `npm ci` and syntax checks pass. A real SDK stdio client initializes,
  lists five tool schemas, calls a tool through a disposable CLI fixture,
  verifies JSON/read-only/no-input flags, resolves the prompt and checks the
  missing-binary error. The real Go binary and provider APIs are never called.
- Full and production dependency audits are zero for both packages; outdated
  checks report no direct updates. There is no configured Node lint suite.

Go code, dependencies, source installation, version files and release workflow
are unchanged; Go tests were not rerun for this Node-only change. CI retains its
original triggers and Go checks. Binary release remains gated by `v*` tags;
docs deployment remains an explicit Wrangler action. No push, release tag,
publish, provider call or deployment was performed.

Webpack reports dynamic-import cache-invalidation warnings from Fumadocs and
CSS highlight-selector optimizer warnings. They do not fail types, export,
search or browser checks and were not suppressed.

The demonstrated gains are Node 26 compatibility and a current SDK/Fumadocs
patch level, with separate package contracts maintained. The existing
zero-advisory dependency state is preserved. No runtime speedup is claimed:
the Go CLI is unchanged and no comparative benchmark was run.
