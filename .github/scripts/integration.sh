#!/usr/bin/env bash
# Call through check-source.sh until the matching composition modules publish.
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
for required in SERVEX_EXAMPLES_IT_POSTGRES_DSN SERVEX_EXAMPLES_IT_REDIS_ADDR \
  SERVEX_EXAMPLES_IT_S3_ENDPOINT SERVEX_EXAMPLES_IT_S3_BUCKET \
  SERVEX_EXAMPLES_IT_S3_ACCESS_KEY SERVEX_EXAMPLES_IT_S3_SECRET_KEY \
  SERVEX_EXAMPLES_IT_ETCD_ENDPOINTS; do
  if [ -z "${!required:-}" ]; then
    printf 'Required variable is empty: %s\n' "$required" >&2
    exit 1
  fi
done
cd "$root"
result="$(mktemp)"
trap 'rm -f "$result"' EXIT
status=0
GOWORK=off "${GO:-go}" test -race -shuffle=on -count=1 -timeout=5m -tags integration \
  -run '^TestIntegration' -json ./... | tee "$result" || status=$?
GOWORK=off "${GO:-go}" run "$root/.github/scripts/integrationcheck/main.go" "$result"
exit "$status"
