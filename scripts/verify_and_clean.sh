#!/usr/bin/env bash
set -euo pipefail

echo "============================================"
echo "   RESTLESS VERSION VERIFY + REPO CLEAN"
echo "============================================"

# Ensure in repo root
if [ ! -f "go.mod" ]; then
  echo "ERROR: Run from repo root"
  exit 1
fi

# Extract version from source
if [ ! -f "internal/version/version.go" ]; then
  echo "ERROR: internal/version/version.go not found"
  exit 1
fi

SRC_VERSION=$(grep 'Version =' internal/version/version.go | cut -d '"' -f2)

if [ -z "$SRC_VERSION" ]; then
  echo "ERROR: Could not extract version from source"
  exit 1
fi

echo "==> Source version: $SRC_VERSION"

# Clean previous binaries
rm -f restless 2>/dev/null || true
rm -rf dist 2>/dev/null || true
mkdir -p dist

echo "==> gofmt"
gofmt -w .

echo "==> go mod tidy"
go mod tidy

echo "==> Running tests"
go test ./...

echo "==> Building with ldflags"
go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$SRC_VERSION" -o restless ./cmd/restless

echo "==> Checking binary version"
BIN_VERSION=$(./restless --version | awk '{print $2}')

if [ "$BIN_VERSION" != "$SRC_VERSION" ]; then
  echo "❌ VERSION MISMATCH"
  echo "Binary: $BIN_VERSION"
  echo "Source: $SRC_VERSION"
  exit 1
fi

echo "✅ Binary reports correct version"

# Git tag check
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ "$LATEST_TAG" != "$SRC_VERSION" ]; then
  echo "⚠ Git tag ($LATEST_TAG) does not match source version ($SRC_VERSION)"
else
  echo "✅ Git tag matches version"
fi

echo "==> Checking for hardcoded version strings"
if grep -R "$SRC_VERSION" . | grep -v internal/version/version.go >/dev/null 2>&1; then
  echo "⚠ Found additional references to version string:"
  grep -R "$SRC_VERSION" . | grep -v internal/version/version.go
else
  echo "✅ No stray hardcoded version strings"
fi

echo "==> Building cross-platform artifacts"
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$SRC_VERSION" -o dist/restless_linux_amd64 ./cmd/restless
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$SRC_VERSION" -o dist/restless_darwin_amd64 ./cmd/restless
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$SRC_VERSION" -o dist/restless_windows_amd64.exe ./cmd/restless

echo "==> Artifacts:"
ls -lh dist

echo "============================================"
echo "   VERIFY + CLEAN COMPLETE"
echo "============================================"
