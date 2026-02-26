#!/usr/bin/env bash
set -euo pipefail

echo "================================================="
echo "  PROMOTING V2 TO PRIMARY CLI (restless)"
echo "================================================="

# Ensure we are in repo root
if [ ! -f "go.mod" ]; then
  echo "ERROR: go.mod not found. Run from repo root."
  exit 1
fi

# Ensure v2 exists
if [ ! -d "cmd/restless-v2" ]; then
  echo "ERROR: cmd/restless-v2 not found."
  exit 1
fi

echo "==> Archiving old CLI if present"

if [ -d "cmd/restless" ]; then
  rm -rf cmd/restless-legacy 2>/dev/null || true
  mv cmd/restless cmd/restless-legacy
  echo "   old CLI moved to cmd/restless-legacy"
fi

echo "==> Promoting cmd/restless-v2 -> cmd/restless"
mv cmd/restless-v2 cmd/restless

echo "==> Updating Makefile build targets"

if [ -f "Makefile" ]; then
  sed -i 's|./cmd/restless-v2|./cmd/restless|g' Makefile
fi

echo "==> Cleaning dist directory"
rm -rf dist 2>/dev/null || true

echo "==> gofmt"
gofmt -w .

echo "==> go mod tidy"
go mod tidy

echo "==> Running tests"
go test ./...

echo "==> Building primary binary"
go build -o restless ./cmd/restless

echo "==> Cross-platform build (if Makefile supports build-all)"
if grep -q "build-all" Makefile 2>/dev/null; then
  make build-all
fi

echo "================================================="
echo "  SUCCESS: v2 is now primary CLI"
echo "================================================="
