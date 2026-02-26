#!/usr/bin/env bash
set -euo pipefail

PKG="restless"
REPO="bspippi1337/restless"
TAP_REPO="bspippi1337/homebrew-restless"
PAGES_URL="https://bspippi1337.github.io/restless/"

need() { command -v "$1" >/dev/null 2>&1 || { echo "❌ Missing: $1"; exit 2; }; }

need git
need go
need tar
need shasum

if command -v gh >/dev/null 2>&1; then
  HAVE_GH=1
else
  HAVE_GH=0
fi

VERSION="$(git describe --tags --always --dirty || echo 0.0.0)"
TAG="v${VERSION}"

echo "🧪 Running tests"
CGO_ENABLED=0 go test ./...

echo "🛠 Building binaries"
mkdir -p dist/releases
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/${PKG}_linux_amd64 ./cmd/restless
tar -czf dist/releases/${PKG}_${VERSION}_linux_amd64.tar.gz -C dist ${PKG}_linux_amd64
SHA_LINUX=$(shasum -a 256 dist/releases/${PKG}_${VERSION}_linux_amd64.tar.gz | awk '{print $1}')

echo "🧠 Updating MANUAL.md"

HELP_OUTPUT="$(./dist/${PKG}_linux_amd64 --help 2>/dev/null || true)"

cat > MANUAL.md <<MANUAL
# Restless Manual

Version: ${VERSION}

## CLI

\`\`\`
${HELP_OUTPUT}
\`\`\`

## Smart Commands
- probe <url>
- smart <url>
- simulate <url>
- export --format=<json|md|curl|har>

## Interactive Mode
Run without arguments:
\`\`\`
restless
\`\`\`

## Install via APT
\`\`\`
echo "deb [trusted=yes] ${PAGES_URL} ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update && sudo apt install restless
\`\`\`

## Install via Homebrew
\`\`\`
brew tap ${TAP_REPO}
brew install restless
\`\`\`
MANUAL

echo "📦 Commit & push if needed"
if [ -n "$(git status --porcelain)" ]; then
  git add -A
  git commit -m "release: ${VERSION} (auto build + manual update)"
  git push
fi

echo "🏷 Ensuring tag exists"
git tag "${TAG}" >/dev/null 2>&1 || true
git push --tags >/dev/null 2>&1 || true

if [ "$HAVE_GH" = "1" ]; then
  echo "📦 Publishing GitHub Release"
  gh release view "${TAG}" >/dev/null 2>&1 || \
    gh release create "${TAG}" -t "${TAG}" -n "Restless ${VERSION}"

  gh release upload "${TAG}" dist/releases/${PKG}_${VERSION}_linux_amd64.tar.gz --clobber

  echo "🍺 Ensuring Homebrew tap exists"
  if ! gh repo view "${TAP_REPO}" >/dev/null 2>&1; then
    gh repo create "${TAP_REPO}" --public --clone
  fi

  TAP_DIR="$(mktemp -d)"
  git clone "https://github.com/${TAP_REPO}.git" "$TAP_DIR" >/dev/null 2>&1 || true
  mkdir -p "${TAP_DIR}/Formula"

  RELEASE_URL="https://github.com/${REPO}/releases/download/${TAG}/${PKG}_${VERSION}_linux_amd64.tar.gz"

  cat > "${TAP_DIR}/Formula/restless.rb" <<RUBY
class Restless < Formula
  desc "Restless adaptive API client"
  homepage "https://github.com/${REPO}"
  version "${VERSION}"

  on_linux do
    url "${RELEASE_URL}"
    sha256 "${SHA_LINUX}"
    def install
      bin.install "${PKG}_linux_amd64" => "restless"
    end
  end

  test do
    system "#{bin}/restless", "--help"
  end
end
RUBY

  cd "$TAP_DIR"
  git add -A
  git commit -m "restless ${VERSION}" >/dev/null || true
  git push
  cd - >/dev/null

  echo "🍺 Brew updated"
fi

echo ""
echo "🚀 AUTOPILOT COMPLETE"
echo "APT published via GitHub Actions."
echo "Brew tap ensured."
echo "Manual updated."
