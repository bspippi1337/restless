#!/usr/bin/env bash
set -euo pipefail

FILE="cmd/restless/openapi_cli.go"

echo "==> Fixing OpenAPI CLI loader (use disk spec, not RawPath)"

if [ ! -f "$FILE" ]; then
  echo "ERROR: $FILE not found"
  exit 1
fi

cp "$FILE" "$FILE.bak"

# Replace LoadSpecFromFile(idx.RawPath) with LoadSpec(ra.ID)
sed -i 's/openapi.LoadSpecFromFile(idx.RawPath)/openapi.LoadSpec(ra.ID)/' "$FILE"

echo "==> Formatting"
go fmt ./...

echo "==> Building"
go build -o restless ./cmd/restless

echo "==> Testing"
go test ./...

echo "==> Commit"
git add "$FILE"
git commit -m "fix(openapi): load spec from disk using ID instead of RawPath"

echo "Done."
