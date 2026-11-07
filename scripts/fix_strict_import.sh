#!/usr/bin/env bash
set -euo pipefail

FILE="internal/modules/openapi/run.go"

echo "==> Fixing missing os import in run.go"

if ! grep -q '"os"' "$FILE"; then
  sed -i '/^import (/a\	"os"' "$FILE"
fi

gofmt -w "$FILE"
go build -o restless-v2 ./cmd/restless-v2

echo "✅ strict import fixed"
