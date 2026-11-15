#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../web"
pnpm install --frozen-lockfile
pnpm build:cloudflare
pnpm test:search
pnpm deploy:cloudflare
