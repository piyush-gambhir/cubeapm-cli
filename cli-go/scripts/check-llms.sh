#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
bin="$(mktemp)"; trap 'rm -f "$bin"' EXIT
go build -o "$bin" .
llms="../docs/llms.txt"
[ -f "$llms" ] || { echo "missing $llms"; exit 1; }
missing=0
while read -r c; do
  [ -z "$c" ] && continue
  case "$c" in help|completion) continue;; esac
  grep -qw -- "$c" "$llms" || { echo "MISSING from $llms: command '$c'"; missing=$((missing+1)); }
done < <("$bin" --help 2>/dev/null | awk '/Available Commands:/{f=1;next} /^Flags:|^Global Flags:/{f=0} f && /^[ ]+[a-z]/{print $1}')
if [ "$missing" -ne 0 ]; then echo "FAIL: $missing command(s) undocumented in $llms"; exit 1; fi
echo "OK: all top-level commands documented in $llms"
