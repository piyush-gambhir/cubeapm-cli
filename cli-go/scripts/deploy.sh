#!/usr/bin/env bash
# Local release wrapper for grafana-cli. Wraps goreleaser.
#
# Usage:
#   bash scripts/deploy.sh production    # full release; requires a vN.N.N tag at HEAD
#   bash scripts/deploy.sh development   # snapshot build, no publish
#
# The actual goreleaser config is in .goreleaser.yaml.
# In CI, .github/workflows/release.yml runs the same goreleaser command on tag push.

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
fi

if ! command -v goreleaser >/dev/null 2>&1; then
  echo "error: goreleaser not installed. Install: https://goreleaser.com/install/" >&2
  exit 1
fi

if [[ "$ENV" == "production" ]]; then
  : "${GITHUB_TOKEN:?set GITHUB_TOKEN in ${DEPLOY_ENV_FILE} (PAT with repo + write:packages)}"

  if ! git describe --tags --exact-match HEAD >/dev/null 2>&1; then
    echo "error: production release requires a tag at HEAD." >&2
    echo "       e.g.  git tag v0.1.0 && git push origin v0.1.0" >&2
    exit 1
  fi

  export GITHUB_TOKEN
  goreleaser release --clean
else
  echo "==> Snapshot build (no publish)"
  goreleaser release --snapshot --clean --skip=publish
fi
