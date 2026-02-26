#!/usr/bin/env bash
set -euo pipefail

VERSION="4.0.4"
BINARY="restless"
PKG="github.com/bspippi1337/restless/internal/version"

echo "======================================"
echo "  Releasing Restless v$VERSION"
echo "======================================"

# Ensure gh exists
if ! command -v gh >/dev/null 2>&1; then
  echo "ERROR: GitHub CLI (gh) not installed."
  exit 1
fi

# Clean
rm -rf dist
mkdir -p dist

echo "==> Formatting"
gofmt -w .

echo "==> Tests"
go test ./...

echo "==> Build (cross-platform)"
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "-X $PKG.Version=$VERSION" -o dist/restless_linux_amd64   ./cmd/restless
CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags "-X $PKG.Version=$VERSION" -o dist/restless_linux_arm64   ./cmd/restless
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-X $PKG.Version=$VERSION" -o dist/restless_darwin_amd64  ./cmd/restless
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "-X $PKG.Version=$VERSION" -o dist/restless_darwin_arm64  ./cmd/restless
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X $PKG.Version=$VERSION" -o dist/restless_windows_amd64.exe ./cmd/restless

echo "==> Generate checksums"
cd dist
sha256sum * > checksums.txt
cd ..

echo "==> Verify version"
go build -ldflags "-X $PKG.Version=$VERSION" -o restless ./cmd/restless
./restless --version

echo "==> Commit (if needed)"
git add -A
git commit -m "release: v$VERSION" || true

echo "==> Tag"
git tag -f v$VERSION
git push origin v$VERSION --force

echo "==> Creating GitHub Release"
gh release create v$VERSION \
  dist/* \
  --title "Restless v$VERSION" \
  --notes "
Restless v$VERSION

Major highlights:
- OpenAPI-first execution engine
- Profiles + session templating
- Interactive param prompt
- Strict mode for CI
- Bench with latency histogram
- Cross-platform static builds

Terminal-first API Workbench.
"

echo "======================================"
echo "  v$VERSION published successfully"
echo "======================================"
