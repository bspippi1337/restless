#!/usr/bin/env bash
set -euo pipefail

echo "==> Preparing v2.0.0 release"

# Remove legacy CLI from primary surface (optional keep in archive)
if [ -d "cmd/restless-legacy" ]; then
  mkdir -p archive
  mv cmd/restless-legacy archive/
  echo "   moved legacy CLI to archive/"
fi

# Clean dist
rm -rf dist || true
mkdir -p dist

echo "==> Final formatting"
gofmt -w .

echo "==> go mod tidy"
go mod tidy

echo "==> Running full test suite"
go test ./...

echo "==> Building primary binary"
go build -o dist/restless ./cmd/restless

echo "==> Cross-platform builds"
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o dist/restless_linux_amd64   ./cmd/restless
CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -o dist/restless_linux_arm64   ./cmd/restless
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o dist/restless_darwin_amd64  ./cmd/restless
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o dist/restless_darwin_arm64  ./cmd/restless
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/restless_windows_amd64.exe ./cmd/restless

echo "==> Release artifacts in dist/"
ls -lh dist

echo "✅ Release prep complete"
