#!/usr/bin/env bash
set -euo pipefail

PKG="restless"
REPO="bspippi1337/restless"

need() { command -v "$1" >/dev/null 2>&1 || { echo "❌ Missing: $1"; exit 2; }; }

echo "🔍 Checking tools..."
need git
need go
need tar
need shasum

if command -v gh >/dev/null 2>&1; then
  HAVE_GH=1
else
  HAVE_GH=0
fi

echo "🧪 Running tests (CGO disabled)"
CGO_ENABLED=0 go test ./...

echo "🛠 Building binary (CGO disabled)"
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/${PKG}_linux_amd64 ./cmd/restless

echo "🧼 Formatting"
go fmt ./...

echo "📦 Checking for changes"
if [ -n "$(git status --porcelain)" ]; then
  git add -A
  git commit -m "feat: smart adaptive api client (build verified)"
  git push
  echo "🚀 Code pushed to GitHub (CI will publish APT)"
else
  echo "✅ No code changes"
fi

VERSION="$(git describe --tags --always --dirty)"
TAG="v${VERSION}"

echo "🏷 Ensuring tag exists"
git tag "${TAG}" >/dev/null 2>&1 || true
git push --tags >/dev/null 2>&1 || true

if [ "$HAVE_GH" = "1" ]; then
  echo "📦 Creating GitHub Release (optional brew use)"
  mkdir -p dist/releases
  TGZ="dist/releases/${PKG}_${VERSION}_linux_amd64.tar.gz"
  tar -czf "$TGZ" -C dist "${PKG}_linux_amd64"
  SHA="$(shasum -a 256 "$TGZ" | awk "{print \$1}")"

  gh release view "${TAG}" >/dev/null 2>&1 || \
    gh release create "${TAG}" -t "${TAG}" -n "Restless ${VERSION}"

  gh release upload "${TAG}" "$TGZ" --clobber

  echo "✅ Release published"
  echo "SHA256: $SHA"
else
  echo "⚠️ gh CLI not installed — skipping GitHub Release"
fi

echo ""
echo "🎉 Done."
echo "APT will auto-publish via GitHub Actions."
echo "Install with:"
echo "  echo \"deb [trusted=yes] https://bspippi1337.github.io/restless/ ./\" | sudo tee /etc/apt/sources.list.d/restless.list"
echo "  sudo apt update && sudo apt install restless"
