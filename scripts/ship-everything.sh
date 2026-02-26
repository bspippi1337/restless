#!/usr/bin/env bash
set -euo pipefail

PKG="restless"
REPO="bspippi1337/restless"
TAP_REPO="${TAP_REPO:-bspippi1337/homebrew-restless}"
PAGES_URL="https://bspippi1337.github.io/restless/"
APT_BRANCH="gh-pages"
MAIN_BRANCH="${MAIN_BRANCH:-main}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "❌ Missing: $1"; exit 2; }; }
need git
need go
need tar
need shasum
need curl

if command -v gh >/dev/null 2>&1; then
  HAVE_GH=1
else
  HAVE_GH=0
fi

if [ "$HAVE_GH" != "1" ]; then
  echo "❌ Missing: gh"
  echo "Install gh then re-run. (Termux: pkg install gh)"
  exit 2
fi

echo "🔍 Repo sanity..."
git rev-parse --is-inside-work-tree >/dev/null
git remote get-url origin >/dev/null

# ---------------------------
# Ensure APT GitHub Actions workflow exists (flat repo, no reprepro)
# ---------------------------
echo "🧩 Ensuring APT workflow exists..."
mkdir -p .github/workflows
if [ ! -f .github/workflows/apt-pages.yml ]; then
cat > .github/workflows/apt-pages.yml <<YAML
name: Publish APT (flat) to GitHub Pages

on:
  push:
    branches: ["${MAIN_BRANCH}"]
  workflow_dispatch: {}

permissions:
  contents: write

jobs:
  apt:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install tooling
        run: |
          sudo apt-get update
          sudo apt-get install -y dpkg-dev

      - name: Build binary
        run: |
          mkdir -p dist
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/restless ./cmd/restless
          chmod +x dist/restless

      - name: Build .deb
        run: |
          VERSION="\${GITHUB_SHA::7}"
          mkdir -p pkg/DEBIAN pkg/usr/bin
          install -m 0755 dist/restless pkg/usr/bin/restless
          cat > pkg/DEBIAN/control <<CTL
Package: restless
Version: 0.0.0+\${VERSION}
Section: utils
Priority: optional
Architecture: amd64
Maintainer: bspippi1337 <noreply@github.com>
Description: Restless universal API client
CTL
          dpkg-deb --build pkg dist/restless_\${VERSION}_amd64.deb
          rm -rf pkg

      - name: Create flat APT repo
        run: |
          mkdir -p apt-repo
          cp dist/*.deb apt-repo/
          cd apt-repo
          dpkg-scanpackages . /dev/null > Packages
          gzip -k -f Packages
          cat > Release <<REL
Origin: restless
Label: restless APT
Suite: stable
Codename: stable
Architectures: amd64
Components: main
Description: APT repo for restless
REL
          cd ..
          touch apt-repo/.nojekyll

      - name: Publish gh-pages
        run: |
          git config user.name github-actions
          git config user.email github-actions@github.com
          git fetch origin ${APT_BRANCH} || true
          git checkout ${APT_BRANCH} || git checkout --orphan ${APT_BRANCH}
          rm -rf *
          cp -a apt-repo/. .
          git add -A
          git commit -m "APT publish" || true
          git push origin ${APT_BRANCH}
YAML
fi

# ---------------------------
# Build + test
# ---------------------------
echo "🧪 Tests (CGO disabled)"
CGO_ENABLED=0 go test ./...

echo "🛠 Build linux amd64 (for manual + release asset)"
mkdir -p dist/releases
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/${PKG}_linux_amd64 ./cmd/restless

# ---------------------------
# Version/tag normalization
# ---------------------------
# Prefer semver tag if present, else fallback to short SHA.
# ALWAYS strip leading v, and NEVER include -dirty in releases.
RAW="$(git describe --tags --abbrev=0 2>/dev/null || true)"
if [ -z "$RAW" ]; then
  RAW="$(git rev-parse --short HEAD)"
fi
CLEAN_VERSION="$(echo "$RAW" | sed 's/^v//' | sed 's/-dirty//')"
TAG="v${CLEAN_VERSION}"

ASSET="${PKG}_${CLEAN_VERSION}_linux_amd64.tar.gz"
RELEASE_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

echo "🏷 Version: ${CLEAN_VERSION}"
echo "🏷 Tag:     ${TAG}"
echo "📦 Asset:   ${ASSET}"

# ---------------------------
# Update MANUAL.md
# ---------------------------
echo "🧠 Updating MANUAL.md"
HELP_OUTPUT="$(./dist/${PKG}_linux_amd64 --help 2>/dev/null || true)"

cat > MANUAL.md <<MANUAL
# Restless Manual

Version: ${CLEAN_VERSION}

## CLI Help

\`\`\`
${HELP_OUTPUT}
\`\`\`

## Smart Commands

- probe <url>        : inspect headers/method hints
- smart <url>        : profile + guided flow (expands over time)
- simulate <url>     : interactive request builder
- export ...         : export helpers (formats vary by build)

## Interactive Mode

Run without args:

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

# ---------------------------
# Update README.md (illustrations + demos)
# ---------------------------
echo "🎨 Updating README.md with illustrations + demos"
cat > README.md <<README
# Restless

\`\`\`
██████╗ ███████╗███████╗████████╗██╗     ███████╗███████╗███████╗
██╔══██╗██╔════╝██╔════╝╚══██╔══╝██║     ██╔════╝██╔════╝██╔════╝
██████╔╝█████╗  ███████╗   ██║   ██║     █████╗  ███████╗███████╗
██╔══██╗██╔══╝  ╚════██║   ██║   ██║     ██╔══╝  ╚════██║╚════██║
██║  ██║███████╗███████║   ██║   ███████╗███████╗███████║███████║
╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   ╚══════╝╚══════╝╚══════╝╚══════╝
\`\`\`

Adaptive API client with interactive mode, probing, simulation and export helpers.

Version: **${CLEAN_VERSION}**

---

## Install

### Debian / Ubuntu (APT)

\`\`\`bash
echo "deb [trusted=yes] ${PAGES_URL} ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update
sudo apt install restless
\`\`\`

### Homebrew (tap)

\`\`\`bash
brew tap ${TAP_REPO}
brew install restless
\`\`\`

---

## Demonstrations

### Interactive Mode (no args)

\`\`\`bash
\$ restless
\`\`\`

\`\`\`
🌀 Restless Interactive Mode
> probe https://api.example.com
> simulate https://api.example.com
> quit
\`\`\`

### Probe

\`\`\`bash
\$ restless probe https://api.example.com
\`\`\`

\`\`\`json
{
  "url": "https://api.example.com",
  "methods": ["GET, POST"],
  "content_types": ["application/json"],
  "discovered_at": "2026-02-23T18:22:00Z"
}
\`\`\`

### Simulate

\`\`\`bash
\$ restless simulate https://api.example.com
\`\`\`

\`\`\`
Method [GET]: POST
URL [https://api.example.com]:
Body: {"name":"pippi"}
\`\`\`

### Smart Mode (guided)

\`\`\`bash
\$ restless smart https://api.example.com
\`\`\`

\`\`\`
[ Probing... ] ██████████ 100%
[ Profiling... ] ████████░░ 80%
[ Suggesting tools... ] ██████████ 100%
Ready.
\`\`\`

### Export (examples)

\`\`\`bash
\$ restless export --format=har --out=req.har get https://api.example.com
\$ restless export --format=curl --out=req.sh  post https://api.example.com
\`\`\`

---

## Architecture

\`\`\`
smartcmd
  │
  ├─ discover  → engine (suggest tools)
  ├─ simulator → guided request builder
  ├─ export    → json/md/har/curl outputs
  └─ app       → existing core (API, fuzzing, etc.)
\`\`\`

---

## Distribution

- **APT**: published automatically to GitHub Pages via Actions (flat repo)
- **Brew**: formula auto-updated in tap repo from GitHub Release assets

Doctor:

\`\`\`bash
./scripts/dist-doctor.sh
\`\`\`

Release asset URL pattern:

\`\`\`
${RELEASE_URL}
\`\`\`

README

# ---------------------------
# gofmt
# ---------------------------
echo "🧼 gofmt"
go fmt ./...

# ---------------------------
# Commit/push (only if changes)
# ---------------------------
echo "📦 Commit & push changes (if any)"
if [ -n "$(git status --porcelain)" ]; then
  git add -A
  git commit -m "docs+ci: refresh README/MANUAL + ensure APT workflow"
  git push
else
  echo "✅ No changes to commit"
fi

# ---------------------------
# Tag + Release + Asset upload
# ---------------------------
echo "🏷 Ensuring tag ${TAG}"
git tag "${TAG}" >/dev/null 2>&1 || true
git push --tags >/dev/null 2>&1 || true

echo "📦 Ensuring GitHub Release ${TAG}"
gh release view "${TAG}" >/dev/null 2>&1 || gh release create "${TAG}" -t "${TAG}" -n "Restless ${CLEAN_VERSION}"

echo "📦 Building release asset"
tar -czf "dist/releases/${ASSET}" -C dist "${PKG}_linux_amd64"
TMPFILE="$(mktemp)"
cp "dist/releases/${ASSET}" "$TMPFILE"
SHA="$(shasum -a 256 "$TMPFILE" | awk "{print \$1}")"
rm -f "$TMPFILE"

echo "⬆️ Uploading asset to release"
gh release upload "${TAG}" "dist/releases/${ASSET}" --clobber

# ---------------------------
# Ensure tap repo exists + update formula
# ---------------------------
echo "🍺 Ensuring Homebrew tap exists: ${TAP_REPO}"
if ! gh repo view "${TAP_REPO}" >/dev/null 2>&1; then
  gh repo create "${TAP_REPO}" --public
fi

TAP_DIR="$(mktemp -d)"
git clone "https://github.com/${TAP_REPO}.git" "$TAP_DIR" >/dev/null 2>&1 || true
mkdir -p "${TAP_DIR}/Formula"

cat > "${TAP_DIR}/Formula/restless.rb" <<RUBY
class Restless < Formula
  desc "Restless adaptive API client"
  homepage "https://github.com/${REPO}"
  version "${CLEAN_VERSION}"

  on_linux do
    url "${RELEASE_URL}"
    sha256 "${SHA}"
    def install
      bin.install "restless_linux_amd64" => "restless"
    end
  end

  test do
    system "#{bin}/restless", "--help"
  end
end
RUBY

(
  cd "$TAP_DIR"
  git add -A
  git commit -m "restless ${CLEAN_VERSION}" >/dev/null || true
  git push
)

echo ""
echo "✅ DONE"
echo "APT (CI) will publish to: ${PAGES_URL}"
echo "Brew tap updated: ${TAP_REPO}"
echo "Release asset URL:"
echo "  ${RELEASE_URL}"
echo ""
echo "Install APT:"
echo "  echo \"deb [trusted=yes] ${PAGES_URL} ./\" | sudo tee /etc/apt/sources.list.d/restless.list"
echo "  sudo apt update && sudo apt install restless"
echo ""
echo "Install Brew:"
echo "  brew tap ${TAP_REPO}"
echo "  brew install restless"
