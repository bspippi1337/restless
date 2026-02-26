#!/usr/bin/env bash
set -euo pipefail

FILE="cmd/restless-v2/openapi_cli.go"

echo "==> Ensuring term import exists correctly"

# Hvis import-blokk finnes, legg til term der
if grep -q '^import (' "$FILE"; then
  if ! grep -q 'internal/ui/term' "$FILE"; then
    sed -i '/^import (/a\
\t"github.com/bspippi1337/restless/internal/ui/term"' "$FILE"
  fi
else
  # fallback: lag ny import-blokk øverst etter package
  sed -i '0,/^package .*/a\
import (\n\t"github.com/bspippi1337/restless/internal/ui/term"\n)' "$FILE"
fi

gofmt -w "$FILE"
go test ./...
go build -o restless-v2 ./cmd/restless-v2

echo "✅ term import fixed"
