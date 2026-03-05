#!/usr/bin/env bash
set -euo pipefail

echo "======================================"
echo " Enforcing single binary: restless"
echo "======================================"

# Remove stray binaries
rm -f restless2 restless-v2 2>/dev/null || true
rm -f dist/restless2* 2>/dev/null || true
rm -f dist/restless-v2* 2>/dev/null || true

# Remove old cmd folder if it exists
if [ -d "cmd/restless-v2" ]; then
  echo "Removing old cmd/restless-v2"
  rm -rf cmd/restless-v2
fi

# Clean Makefile
if [ -f "Makefile" ]; then
  echo "Cleaning Makefile..."
  sed -i '/restless-v2/d' Makefile
  sed -i '/restless2/d' Makefile
fi

# Clean dist
rm -rf dist
mkdir -p dist

echo "==> gofmt"
gofmt -w .

echo "==> go mod tidy"
go mod tidy

echo "==> test"
go test ./...

echo "==> building primary binary"
go build -o restless ./cmd/restless

echo "==> cross build"
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o dist/restless_linux_amd64   ./cmd/restless
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o dist/restless_darwin_amd64  ./cmd/restless
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/restless_windows_amd64.exe ./cmd/restless

echo
echo "======================================"
echo " Only binary now: restless"
echo "======================================"

ls -lh restless
ls -lh dist
