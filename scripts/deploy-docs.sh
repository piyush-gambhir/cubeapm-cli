#!/usr/bin/env bash
# Build the Next.js site in web/ and deploy the static export to Cloudflare Pages.
# Run as `bash scripts/deploy-docs.sh [production|development]`.
#
# Falls back to a local `wrangler login` session if no .env.deploy.<env> file is present.
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
  if ! npx --yes wrangler@4.110.0 whoami >/dev/null 2>&1; then
    echo "error: not logged in to wrangler. Run \`wrangler login\` first (or create $DEPLOY_ENV_FILE)." >&2
    exit 1
  fi
fi

CF_PROJECT_NAME="${CF_PROJECT_NAME:-cubeapm-cli}"
WEB_DIR="${WEB_DIR:-web}"
OUT_DIR="${OUT_DIR:-$WEB_DIR/out}"

echo "==> Building the site in ${WEB_DIR}/"
( cd "$WEB_DIR" && pnpm install --frozen-lockfile && pnpm build )

if [[ ! -f "$OUT_DIR/index.html" ]]; then
  echo "error: $OUT_DIR/index.html not found — build produced no static export." >&2
  exit 1
fi

if [[ "$ENV" == "production" ]]; then
  CF_BRANCH="${CF_PRODUCTION_BRANCH:-main}"
else
  CF_BRANCH="${CF_PREVIEW_BRANCH:-preview}"
fi

if [[ -n "${CLOUDFLARE_API_TOKEN:-}" ]]; then export CLOUDFLARE_API_TOKEN; fi
if [[ -n "${CLOUDFLARE_ACCOUNT_ID:-}" ]]; then export CLOUDFLARE_ACCOUNT_ID; fi

echo "==> Deploying ${OUT_DIR}/ to Cloudflare Pages project '${CF_PROJECT_NAME}' (branch: ${CF_BRANCH})"
npx --yes wrangler@4.110.0 pages deploy "$OUT_DIR" \
  --project-name="$CF_PROJECT_NAME" \
  --branch="$CF_BRANCH"
