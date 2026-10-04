#!/usr/bin/env bash
# Builds the docs site in web/ and deploys it as the Cloudflare Worker that serves
# projects.piyushgambhir.com/cubeapm-cli. Run from an up-to-date main after
# `wrangler login` (or with CLOUDFLARE_API_TOKEN set).
set -euo pipefail

cd "$(dirname "$0")/../web"
pnpm install --frozen-lockfile
pnpm build:cloudflare
pnpm test:search
pnpm deploy:cloudflare
