#!/usr/bin/env bash
# Local Cloudflare Pages deploy for the docs/ landing page.
# Run as `bash scripts/deploy-docs.sh [production|development]`.
#
# Falls back to local `wrangler login` session if no .env.deploy.<env> file is present.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

ENV="${1:-production}"
DEPLOY_ENV_FILE=".env.deploy.${ENV}"

if [[ -f "$DEPLOY_ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$DEPLOY_ENV_FILE"
  set +a
else
  if ! npx --yes wrangler@latest whoami >/dev/null 2>&1; then
    echo "error: not logged in to wrangler. Run \`wrangler login\` first." >&2
    exit 1
  fi
fi

CF_PROJECT_NAME="${CF_PROJECT_NAME:-cubeapm-cli}"
DOCS_DIR="${DOCS_DIR:-docs}"

if [[ ! -f "$DOCS_DIR/index.html" ]]; then
  echo "error: $DOCS_DIR/index.html not found — nothing to deploy." >&2
  exit 1
fi

if [[ "$ENV" == "production" ]]; then
  CF_BRANCH="${CF_PRODUCTION_BRANCH:-main}"
else
  CF_BRANCH="${CF_PREVIEW_BRANCH:-preview}"
fi

if [[ -n "${CLOUDFLARE_API_TOKEN:-}" ]]; then export CLOUDFLARE_API_TOKEN; fi
if [[ -n "${CLOUDFLARE_ACCOUNT_ID:-}" ]]; then export CLOUDFLARE_ACCOUNT_ID; fi

echo "==> Deploying ${DOCS_DIR}/ to Cloudflare Pages project '${CF_PROJECT_NAME}' (branch: ${CF_BRANCH})"
npx --yes wrangler@latest pages deploy "$DOCS_DIR" \
  --project-name="$CF_PROJECT_NAME" \
  --branch="$CF_BRANCH"
