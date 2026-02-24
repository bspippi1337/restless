#!/usr/bin/env bash
set -euo pipefail

echo "==> Promoting restless-v2 to primary CLI"

# Safety check
if [ ! -d "cmd/restless-v2" ]; then
  echo "ERROR: cmd/restless-v2 not found"
  exit 1
fi

# Backup old CLI if exists
if [ -d "cmd/restless" ]; then
  echo "==> Archiving old CLI to cmd/restless-legacy"
  rm -rf cmd/restless-legacy || true
  mv cmd/restless cmd/restless-legacy
fi

# Promote v2
mv cmd/restless-v2 cmd/restless

echo "==> Updating Makefile build targets"

# Replace build targets to point to cmd/restless
sed -i 's|./cmd/restless-v2|./cmd/restless|g' Makefile || true

echo "==> Formatting"
gofmt -w .

echo "==> Testing"
go test ./...

echo "==> Building"
go build -o restless ./cmd/restless

echo "✅ v2 is now primary CLI"
