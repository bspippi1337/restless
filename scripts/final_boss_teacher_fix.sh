#!/usr/bin/env bash
set -euo pipefail

FILE="cmd/restless/openapi_cli.go"

echo "================================================="
echo " FINAL BOSS: Fix openapi run loader + teacher"
echo "================================================="

if [ ! -f "$FILE" ]; then
  echo "ERROR: $FILE not found"
  exit 1
fi

echo "==> Restoring $FILE to clean state (discarding broken patches)"
git restore "$FILE" 2>/dev/null || git checkout -- "$FILE"

echo "==> Patching openapi run: stop trusting idx.RawPath, load spec from disk via ID"

# Replace ONLY the spec-loading block in case "run":
# We match the exact sequence:
#   idx, err := openapi.LoadIndex(ra.ID)
#   ...
#   spec, err := openapi.LoadSpecFromFile(idx.RawPath)
#
# Then we replace the spec-load line with deterministic disk load.
perl -0777 -i -pe '
s{
(\t\tidx,\s*err\s*:=\s*openapi\.LoadIndex\(ra\.ID\)\s*\n
\t\tif\s*err\s*!=\s*nil\s*\{\s*\n
\t\t\tfmt\.Println\("index error:",\s*err\)\s*\n
\t\t\tos\.Exit\(1\)\s*\n
\t\t\}\s*\n
)
\t\tspec,\s*err\s*:=\s*openapi\.LoadSpecFromFile\(idx\.RawPath\)
\t\tif\s*err\s*!=\s*nil\s*\{\s*\n
\t\t\tfmt\.Println\("ERROR: spec:",\s*err\)\s*\n
\t\t\tos\.Exit\(1\)\s*\n
\t\t\}
}{
$1\t\topenapiDir, err := openapi.DefaultDir()
\t\tif err != nil {
\t\t\tfmt.Println("ERROR:", err)
\t\t\tos.Exit(1)
\t\t}

\t\tspecPath := filepath.Join(openapiDir, ra.ID+".json")

\t\tspec, err := openapi.LoadSpecFromFile(specPath)
\t\tif err != nil {
\t\t\tfmt.Println("ERROR: spec:", err)
\t\t\tos.Exit(1)
\t\t}
}gsx
' "$FILE"

echo "==> Ensuring filepath import exists (in the main import block)"
if ! grep -q '"path/filepath"' "$FILE"; then
  # insert after import ( line
  sed -i '/^import (/a\
\t"path/filepath"' "$FILE"
fi

echo "==> Sanity check: ensure no LoadSpec(ra.ID) remains"
if grep -n 'LoadSpec(ra.ID)' "$FILE" >/dev/null 2>&1; then
  echo "ERROR: Found LoadSpec(ra.ID) still present; refusing to continue."
  grep -n 'LoadSpec(ra.ID)' "$FILE" || true
  exit 1
fi

echo "==> gofmt"
go fmt ./...

echo "==> build"
go build -o restless ./cmd/restless

echo "==> teacher"
./restless teacher

echo "==> commit"
git add "$FILE"
git commit -m "fix(openapi): load spec from disk via ID (teacher stable)" || {
  echo "No changes to commit (already applied)."
}

echo "✅ FINAL BOSS DONE"
