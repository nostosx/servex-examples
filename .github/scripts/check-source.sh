#!/usr/bin/env bash
# This unreleased composition consumes matching contrib source explicitly.
# No local replace or private checksum is written to canonical manifests.
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
source_root="${SERVEX_CONTRIB_SOURCE:-$root/../servex-contrib}"
source_root="$(cd "$source_root" && pwd)"
# GOFLAGS is an environment variable, so Git Bash cannot translate its paths
# when starting native Go. Give Go native paths while retaining Bash's PATH.
if command -v cygpath >/dev/null 2>&1; then source_root="$(cygpath -am "$source_root")"; fi
temporary="$(mktemp -d)"
trap 'rm -rf "$temporary"' EXIT
cp "$root/go.mod" "$temporary/source.mod"
if [ -f "$root/go.sum" ]; then cp "$root/go.sum" "$temporary/source.sum"; fi
go_binary="$(command -v "${GO:-go}")"
export SERVEX_SOURCE_GO="$go_binary" SERVEX_SOURCE_MODFILE="$temporary/source.mod" GOWORK=off
if command -v cygpath >/dev/null 2>&1; then SERVEX_SOURCE_MODFILE="$(cygpath -am "$SERVEX_SOURCE_MODFILE")"; fi
export SERVEX_SOURCE_FLAG="'-modfile=$SERVEX_SOURCE_MODFILE'"
replacements=()
for manifest in "$source_root"/*/go.mod "$source_root"/*/*/go.mod; do
  if [ -f "$manifest" ]; then
    directory="${manifest%/go.mod}"
    module="${directory#"$source_root"/}"
    replacements+=("-replace=github.com/nostosx/servex-contrib/$module=$directory")
  fi
done
cd "$root"
"$go_binary" mod edit -modfile="$SERVEX_SOURCE_MODFILE" "${replacements[@]}"
mkdir "$temporary/bin"
cat > "$temporary/bin/go" <<'GO_WRAPPER'
#!/usr/bin/env bash
set -euo pipefail
if [ "${GO111MODULE:-}" = off ]; then export GOFLAGS="${GOFLAGS//"$SERVEX_SOURCE_FLAG"/}"; fi
exec "$SERVEX_SOURCE_GO" "$@"
GO_WRAPPER
chmod +x "$temporary/bin/go"
export PATH="$temporary/bin:$PATH"
export GOFLAGS="${GOFLAGS:-} $SERVEX_SOURCE_FLAG"
if [ "${1:-}" = -- ]; then
  shift
  "$@"
else
  task --silent "$@"
fi
